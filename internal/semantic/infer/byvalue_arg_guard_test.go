package infer_test

import "testing"

// Passing a local to a by-value parameter of a method inside a branch does
// not define it again: an early exit before the branch still narrows it
// after the branch.
func TestByValueMethodArgumentKeepsGuards(t *testing.T) {
	checkAnywhere(t, `<?php
interface Report { public function changes(string $path): ?string; }
class Paths { public function find(): ?string { return null; } }

function typed(Paths $paths, $tool, $flag) {
    $dir = $paths->find();
    if ($dir === null) {
        return;
    }
    if ($tool instanceof Report) {
        $tool->changes($dir);
    }
    t('narrowedReceiver', $dir);
    if ($flag) {
        $tool->changes($dir);
    }
    t('untypedReceiver', $dir);
}
`, map[string]string{"narrowedReceiver": "string", "untypedReceiver": "string"})
}

// A guard between a definition and a later, conditional one still rules
// null out of the earlier value.
func TestGuardBeforeConditionalRedefinition(t *testing.T) {
	checkAnywhere(t, `<?php
function conv(string $u): string { return $u; }
function f(?string $url, $c) {
    if (!$url) {
        return;
    }
    if ($c) {
        $url = conv($url);
    }
    t('truthy', $url);
}
function g(?string $url, $c) {
    if ($url === null) {
        return;
    }
    if ($c) {
        $url = null;
    }
    t('reassignedNull', $url);
}
`, map[string]string{"truthy": "string", "reassignedNull": "null|string"})
}
