<?php
function pause($timeout) {} pause(<error descr="Use a declared parameter name.">timeot:3</error>);

// Verified builtins keep their contract.
strlen(<error descr="Use a declared parameter name.">str: 'a'</error>);
function tick(DateTime $at) { return $at->format(<error descr="Use a declared parameter name.">fmt: 'c'</error>); }

class Renderer { public function render($template) {} }
class TwigRenderer extends Renderer { public function render($view) {} }
// No subtype declares the name either.
function show(Renderer $r) { $r->render(<error descr="Use a declared parameter name.">tpl: 'home'</error>); }
// A named class, self and parent bind to that class's own method; private
// and final methods have no override.
class Fac {
    public static function make($a) {}
    public static function call() { self::make(<error descr="Use a declared parameter name.">b: 1</error>); }
    private function own($a) {}
    public function go() { $this->own(<error descr="Use a declared parameter name.">b: 1</error>); }
    final public function sealed($a) {}
}
class Fac2 extends Fac {
    public static function make($b) {}
    public function own($b) {}
    public static function up() { parent::make(<error descr="Use a declared parameter name.">b: 1</error>); }
}
Fac::make(<error descr="Use a declared parameter name.">b: 1</error>);
function sealed(Fac $f) { $f->sealed(<error descr="Use a declared parameter name.">b: 1</error>); }
final class Leaf { public function m($a) {} }
function leaf(Leaf $l) { $l->m(<error descr="Use a declared parameter name.">b: 1</error>); }
class Plain { public function a($x) {} }
class Other { public function a($y) {} }
function either(Plain|Other $o) { $o->a(<error descr="Use a declared parameter name.">z: 1</error>); }
if (PHP_VERSION_ID > 80000) { function cond($a) {} } else { function cond($b) {} }
cond(<error descr="Use a declared parameter name.">c: 1</error>);
