package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

var (
	labelName       = flag.String("label", "", "Label to clean (required)")
	olderThanSearch = flag.String("older", "", "Gmail-style \"older than\" search string (eg. '1y' for 1 year, '3m' for 3 months) (required)")
	excludeSearch   = flag.String("exclude", "", "Additional Gmail-style search string specifying results to exclude.")
	searchCap       = flag.Int64("cap", 500, "Cap on the number of emails to trash. If the (estimated) result count exceeds this, no data will be modified.")
	actuallyTrash   = flag.Bool("trash", false, "Whether to actually trash discovered threads. By default, no data will be modified.")
)

func Main() error {
	flag.Parse()

	if *labelName == "" {
		return errors.New("argument 'label' is required")
	}
	if strings.Contains(*labelName, "\"") {
		return errors.New("argument 'label' must not contain any double quotes (\")")
	}

	if *olderThanSearch == "" {
		return errors.New("argument 'older' is required")
	}
	olderThanRegex := regexp.MustCompile("\\A\\d+[ymd]\\z")
	if !olderThanRegex.MatchString(*olderThanSearch) {
		return errors.New("argument 'older' must be of the form '\\d[ymd]'")
	}

	if strings.Contains(*excludeSearch, ")") {
		return errors.New("argument 'exclude' must not contain closing parens")
	}

	srv, err := buildGmailService()
	if err != nil {
		return err
	}

	searchQ := fmt.Sprintf("label:\"%s\" older_than:%s", *labelName, *olderThanSearch)
	if *excludeSearch != "" {
		searchQ = fmt.Sprintf("%s -(%s)", searchQ, *excludeSearch)
	}
	log.Printf("search query: \"%s\"\n", searchQ)
	log.Printf("gmail search: https://mail.google.com/mail/#search/%s\n", url.QueryEscape(searchQ))

	ctx := context.Background()
	var threadIds []string

	if err = srv.Users.Threads.List("me").IncludeSpamTrash(false).Q(searchQ).Context(ctx).Pages(ctx, func(response *gmail.ListThreadsResponse) error {
		if response.ResultSizeEstimate > *searchCap {
			return fmt.Errorf("too many results! estimated result count %d is above cap %d", response.ResultSizeEstimate, *searchCap)
		}
		for _, t := range response.Threads {
			threadIds = append(threadIds, t.Id)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error searching for messages to trash: %w", err)
	}

	log.Printf("found %d threads to trash\n", len(threadIds))
	if !*actuallyTrash {
		log.Println("not actually trashing anything (flag -trash is missing)")
	}
	threadsTrashed := 0

	for _, tId := range threadIds {
		t, err := srv.Users.Threads.Get("me", tId).Context(ctx).Do()
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
			_, err = srv.Users.Threads.Trash("me", tId).Context(ctx).Do()
			if err != nil {
				log.Printf("trashed %d threads.\n", threadsTrashed)
				return fmt.Errorf("unable to trash thread \"%s\": %w", subject, err)
			}
			threadsTrashed++
		}
	}

	log.Printf("trashed %d threads.\n", threadsTrashed)
	return nil
}

func main() {
	if err := Main(); err != nil {
		log.Fatalf("%s", err)
	}
}
