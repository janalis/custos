package infer_test

import "testing"

func TestPropertyWritesDropNull(t *testing.T) {
	checkAnywhere(t, `<?php
namespace App;
class Foo { public function go(): void {} }
class Svc {
    private ?Foo $p = null;
    public ?Foo $q;
    public function a() {
        $this->p = new Foo();
        t('assigned', $this->p);
    }
    public function a2() {
        $this->p = new Foo();
        $this->p->go();
        t('afterCall', $this->p);
    }
    public function b() {
        if ($this->p === null) { $this->p = new Foo(); }
        t('lazy', $this->p);
    }
    public function b2() {
        if ($this->p === null) { $this->p = new Foo(); }
        strlen('x');
        t('afterBuiltin', $this->p);
    }
    public function b3() {
        if ($this->p === null) { $this->p = new Foo(); }
        foo();
        t('afterUnknown', $this->p);
    }
    public function c(?Foo $f) {
        $this->p = $f;
        t('nullable', $this->p);
    }
    public function c2() {
        $this->q ??= new Foo();
        t('coalesce', $this->q);
    }
    public function d(bool $c) {
        if ($c) { $this->p = new Foo(); }
        t('branch', $this->p);
    }
    public function d2(bool $c) {
        $this->p = new Foo();
        if ($c) { t('nested', $this->p); }
    }
    public function d3() {
        $this->p = new Foo();
        $x = function () { return $this->p->go(); };
        t('afterClosure', $this->p);
    }
    public function e() {
        $this->p = new Foo();
        $this->other();
        t('afterThisCall', $this->p);
    }
    private function other(): void {}
    public function f() {
        $this->p = new Foo();
        while (true) { t('loop', $this->p); $this->p = null; }
    }
    public function g() {
        $this->p = new Foo();
        $this->p = null;
        t('reset', $this->p);
    }
    public function h($u) {
        $this->p = $u;
        t('unknownValue', $this->p);
        $this->p = &$u;
    }
}
`, map[string]string{
		"assigned": `\App\Foo`, "afterCall": `\App\Foo|null`,
		"lazy": `\App\Foo`, "afterBuiltin": `\App\Foo`, "afterUnknown": `\App\Foo|null`,
		"nullable": `\App\Foo|null`, "coalesce": `\App\Foo`,
		"branch": `\App\Foo|null`, "nested": `\App\Foo`, "afterClosure": `\App\Foo`,
		"afterThisCall": `\App\Foo|null`, "loop": `\App\Foo|null`, "reset": `\App\Foo|null`, "unknownValue": `\App\Foo|null`,
	})
}
