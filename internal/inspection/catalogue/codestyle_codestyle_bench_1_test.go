package catalogue

import (
	"strings"
)

// benchSrc has no findings, so the benchmark measures the Check fast paths.
var benchSrc = "<?php\n" + strings.Repeat(`
class C {
    final public function a() { if (!$x && !($y)) { return !f(); } }
    private function b() { for (;;) { f(); } while (g()) { f(); } }
}
?><p><?= $t ?></p><?php
`, 200)
