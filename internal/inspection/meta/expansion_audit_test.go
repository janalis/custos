package meta

import "testing"

func TestExpansionIntentPoliciesAreOptIn(t *testing.T) {
	for _, id := range []string{"SeekSuccessCheckedAsTruthy", "AsyncSignalSetterReturnMisread", "CollatorComparisonTruthinessReversed", "WaitStatusUsedAsExitCode", "PgAsyncDispatchAssumedQuerySuccess"} {
		r, ok := Lookup(id)
		if !ok || r.EnabledByDefault {
			t.Errorf("%s must be an explicit policy", id)
		}
	}
}
