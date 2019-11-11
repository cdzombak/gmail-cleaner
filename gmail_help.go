package main

import (
	"strings"

	"google.golang.org/api/gmail/v1"
)

func msgSubject(message *gmail.Message) string {
	for _, h := range message.Payload.Headers {
		if strings.ToLower(h.Name) == "subject" {
			return h.Value
		}
	}
	return ""
}
