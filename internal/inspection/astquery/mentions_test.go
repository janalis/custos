package astquery

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestMentions(t *testing.T) {
	f := syntax.Parse("a.php", []byte("<?php use function PCNTL_SigWaitInfo as wait; wait($s);"), syntax.Options{})
	if !Mentions(f, "pcntl_sigwaitinfo") || !Mentions(f, "PCNTL_SIGWAITINFO") || Mentions(f, "pcntl_fork") {
		t.Fatal("mentions")
	}
}
