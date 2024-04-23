package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

var version = "<dev>"

var (
	labelName          = flag.String("label", "", "Label to clean (required)")
	includeSpamTrash   = flag.Bool("include-spam-trash", false, "Whether to include threads in Spam and Trash in the search.")
	olderThanSearch    = flag.String("older", "", "Gmail-style \"older than\" search string (eg. '1y' for 1 year, '3m' for 3 months) (required)")
	excludeSearch      = flag.String("exclude", "", "Additional Gmail-style search string specifying results to exclude.")
	searchCap          = flag.Int64("cap", 500, "Cap on the number of emails to trash. If the (estimated) result count exceeds this, no data will be modified.")
	actuallyTrash      = flag.Bool("trash", false, "Whether to trash discovered threads. By default, no data will be modified.")
	irreversiblyDelete = flag.Bool("irreversibly-delete", false, "Whether to irreversibly delete discovered threads. You should probably use -trash instead. By default, no data will be modified.")
	configDir          = flag.String("configDir", "", "Path to a directory where credentials & user authorization tokens are stored. Overrides environment variable GMAIL_CLEANER_CONFIG_DIR.")
	printVersionFlag   = flag.Bool("version", false, "Print version and exit.")
)

// Main implements the core gmail-cleaner program
func Main() error {
	flag.Parse()

	if *printVersionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *actuallyTrash && *irreversiblyDelete {
		return errors.New("only one of -trash or -irreversibly-delete may be used")
	}

	if *labelName == "" {
		flag.PrintDefaults()
		return errors.New("argument 'label' is required")
	}
	if strings.Contains(*labelName, "\"") {
		return errors.New("argument 'label' must not contain any double quotes (\")")
	}

	if *olderThanSearch == "" {
		flag.PrintDefaults()
		return errors.New("argument 'older' is required")
	}
	olderThanRegex := regexp.MustCompile(`\A\d+[ymd]\z`)
	if !olderThanRegex.MatchString(*olderThanSearch) {
		return errors.New("argument 'older' must be of the form '\\d[ymd]'")
	}

	if strings.Contains(*excludeSearch, "{") || strings.Contains(*excludeSearch, "}") {
		return errors.New("argument 'exclude' must not contain braces ({})")
	}

	if *configDir != "" {
		_ = os.Setenv("GMAIL_CLEANER_CONFIG_DIR", *configDir)
	} else if os.Getenv("GMAIL_CLEANER_CONFIG_DIR") == "" {
		flag.PrintDefaults()
		return errors.New("argument 'configDir' is required (if not using environment variable GMAIL_CLEANER_CONFIG_DIR)")
	}

	srv, err := buildGmailService()
	if err != nil {
		return err
	}

	searchQ := fmt.Sprintf("label:\"%s\" older_than:%s -is:starred", *labelName, *olderThanSearch)
	if *excludeSearch != "" {
		searchQ = fmt.Sprintf("%s -{%s}", searchQ, *excludeSearch)
	}
	log.Printf("search query: \"%s\"\n", searchQ)
	log.Printf("gmail search: https://mail.google.com/mail/#search/%s\n", url.QueryEscape(searchQ))

	ctx := context.Background()
	var threadIDs []string

	if err = srv.Users.Threads.List("me").IncludeSpamTrash(*includeSpamTrash).Q(searchQ).Context(ctx).Pages(ctx, func(response *gmail.ListThreadsResponse) error {
		if response.ResultSizeEstimate > *searchCap {
			return fmt.Errorf("too many results! estimated result count %d is above cap %d", response.ResultSizeEstimate, *searchCap)
		}
		for _, t := range response.Threads {
			threadIDs = append(threadIDs, t.Id)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error searching for messages to trash: %w", err)
	}

	log.Printf("found %d threads to trash\n", len(threadIDs))
	if !*actuallyTrash && !*irreversiblyDelete {
		log.Println("not actually trashing anything (flags -trash or -irreversibly-delete are missing)")
	} else if !*irreversiblyDelete {
		log.Println("matching threads will be irreversibly deleted, not moved to trash (flag -irreversibly-delete is present)")
	}
	threadsTrashed := 0

	for _, tID := range threadIDs {
		t, err := srv.Users.Threads.Get("me", tID).Context(ctx).Do()
		if err != nil {
			return err
		}

		msgCount := 0
		msgTimeMsInt := int64(0)
		subject := ""

		for _, m := range t.Messages {
			msgCount++
			if subject == "" {
				subject = msgSubject(m)
			}
			if m.InternalDate > msgTimeMsInt {
				msgTimeMsInt = m.InternalDate
			}
		}

		if subject == "" {
			subject = t.Snippet
		}

		msgTime := time.Unix(msgTimeMsInt/1000, 0)
		log.Printf("\"%s\" (%s, %d messages)\n", subject, msgTime.Format("2006-01-02"), msgCount)

		if *actuallyTrash {
			_, err = srv.Users.Threads.Trash("me", tID).Context(ctx).Do()
			if err != nil {
				log.Printf("trashed %d threads.\n", threadsTrashed)
				return fmt.Errorf("unable to trash thread \"%s\": %w", subject, err)
			}
			threadsTrashed++
		} else if *irreversiblyDelete {
			err = srv.Users.Threads.Delete("me", tID).Context(ctx).Do()
			if err != nil {
				log.Printf("irreversibly deleted %d threads.\n", threadsTrashed)
				return fmt.Errorf("unable to delete thread \"%s\": %w", subject, err)
			}
			threadsTrashed++
		}
	}

	if *actuallyTrash {
		log.Printf("trashed %d threads.\n", threadsTrashed)
	} else if *irreversiblyDelete {
		log.Printf("irreversibly deleted %d threads.\n", threadsTrashed)
	} else {
		log.Printf("matched %d threads, but did not trash any (flag -trash is not present).\n", threadsTrashed)
	}
	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatalf(err.Error())
	}
}
