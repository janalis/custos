package infer_test

import "testing"

// Inside a trait requiring its users to extend a class (or implement an
// interface), $this is one of them: their members type its calls, after
// the trait's own.
func TestTraitRequiredClass(t *testing.T) {
	check(t, `<?php
class Controller {
    public function user(): ?User { return null; }
    public function flash(): Flash { return new Flash; }
}
interface Named { public function name(): string; }
class User {} class Flash {} class Own {}
/**
 * @phpstan-require-extends Controller
 * @require-implements Named
 */
trait Guard {
    private Own $own;
    public function flash(): Own { return new Own; }
    public function check() {
        t('this', $this);
        t('user', $this->user());
        t('name', $this->name());
        t('flash', $this->flash());
        t('own', $this->own);
        t('missing', $this->missing());
    }
}
trait Plain {
    public function check() { t('plain', $this); }
}
`, map[string]string{
		"this":    `\Controller&\Named`,
		"user":    `\User|null`,
		"name":    "string",
		"flash":   `\Own`,
		"own":     `\Own`,
		"missing": "?unknown",
		"plain":   "?unknown",
	})
}
