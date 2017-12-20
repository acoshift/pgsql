package pgsql

import "github.com/lib/pq"

// IsUniqueViolation checks is error unique_violation with given constraint,
// constraint can be empty to ignore constraint name checks
func IsUniqueViolation(err error, constraint string) bool {
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		if len(constraint) == 0 {
			return true
		}
		return pqErr.Constraint == constraint
	}
	return false
}
