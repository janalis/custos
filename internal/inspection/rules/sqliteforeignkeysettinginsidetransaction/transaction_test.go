package sqliteforeignkeysettinginsidetransaction

import (
	"testing"

	"custos/internal/inspection/semanticquery"
)

// TestTransactionState follows BEGIN, END and ROLLBACK through a script and
// flags a foreign_keys pragma inside a transaction.
func TestTransactionState(t *testing.T) {
	for _, tc := range []struct {
		sql         string
		active, bad bool
	}{
		{";BEGIN;PRAGMA foreign_keys=ON;", true, true},
		{"BEGIN;ROLLBACK;PRAGMA foreign_keys(ON);", false, false},
		{"BEGIN;END;", false, false},
		{"BEGIN;PRAGMA cache_size=100;", true, false},
	} {
		tokens, _ := semanticquery.SQLTokens(tc.sql)
		active, bad := transactionState(tokens, false)
		if active != tc.active || bad != tc.bad {
			t.Fatalf("state %q: %v %v", tc.sql, active, bad)
		}
	}
}
