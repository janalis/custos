package securityadvisories

import (
	"testing"
)

func TestSecurityAdvisoriesFilePatterns(t *testing.T) {
	if got := (securityAdvisories{}).FilePatterns(); len(got) != 1 || got[0] != "composer.json" {
		t.Errorf("FilePatterns() = %v", got)
	}
}
