package syntax

import "testing"

func TestStrictTypes(t *testing.T) {
	for src, want := range map[string]bool{
		`<?php declare(strict_types=1); f();`:     true,
		`<?php declare(STRICT_TYPES=1);`:          true,
		`<?php declare(ticks=1, strict_types=1);`: true,
		`<?php declare(strict_types=0);`:          false,
		`<?php declare(strict_types=ONE);`:        false,
		`<?php declare(ticks=1); f();`:            false,
		`<?php f();`:                              false,
		`<?php declare(strict_types=1) { f(); }`:  true,
	} {
		if got := StrictTypes(Parse("t.php", []byte(src), Options{})); got != want {
			t.Errorf("%s: got %v, want %v", src, got, want)
		}
	}
}
