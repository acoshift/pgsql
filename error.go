package pgsql

import (
	"errors"
	"regexp"

	"github.com/lib/pq"
)

func contains(xs []string, x string) bool {
	for _, p := range xs {
		if p == x {
			return true
		}
	}
	return false
}

// IsUniqueViolation checks is error an unique_violation with given constraint,
// constraint can be empty to ignore constraint name checks
func IsUniqueViolation(err error, constraint ...string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		if len(constraint) == 0 {
			return true
		}
		return contains(constraint, extractConstraint(pqErr))
	}
	return false
}

// IsInvalidTextRepresentation checks is error an invalid_text_representation
func IsInvalidTextRepresentation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "22P02" {
		return true
	}
	return false
}

// IsForeignKeyViolation checks is error an foreign_key_violation
func IsForeignKeyViolation(err error, constraint ...string) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		if len(constraint) == 0 {
			return true
		}
		return contains(constraint, extractConstraint(pqErr))
	}
	return false
}

func extractConstraint(err *pq.Error) string {
	if err.Constraint != "" {
		return err.Constraint
	}
	if err.Message == "" {
		return ""
	}
	if s := extractCRDBKey(err.Message); s != "" {
		return s
	}
	if s := extractLastQuote(err.Message); s != "" {
		return s
	}
	return ""
}

var reLastQuoteExtractor = regexp.MustCompile(`"([^"]*)"[^"]*$`)

// extractLastQuote extracts last string in quote
// ex. `insert or update on table "b" violates foreign key constraint "a_id_fkey"`
// will return `a_id_fkey`
func extractLastQuote(s string) string {
	rs := reLastQuoteExtractor.FindStringSubmatch(s)
	if len(rs) < 2 {
		return ""
	}
	return rs[1]
}

var reCRDBKeyExtractor = regexp.MustCompile(`(\w+@\w+)[^@]*$`)

// extractCRDBKey extracts key from crdb
// until (https://github.com/cockroachdb/cockroach/issues/36494) resolved
// ex. `foreign key violation: value ['b'] not found in a@primary [id] (txn=e3f9af56-5f73-4899-975c-4bb1de800402)`
// will return `a@primary`
func extractCRDBKey(s string) string {
	rs := reCRDBKeyExtractor.FindStringSubmatch(s)
	if len(rs) < 2 {
		return ""
	}
	return rs[1]
}
