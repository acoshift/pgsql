package pgstmt_test

import (
	"strings"
)

func stripSpace(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.TrimSpace(s)
	for {
		p := strings.ReplaceAll(s, "  ", " ")
		if s == p {
			break
		}
		s = p
	}
	return s
}
