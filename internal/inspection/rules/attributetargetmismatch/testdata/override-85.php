<?php
// PHP 8.5 adds properties to Override's targets, not constants.
class P { public $x; const X = 1; public function m() {} }
class C extends P {
    #[\Override] public $x;
    <error descr="Use the attribute on a permitted target.">#[\Override]</error> const X = 2;
    #[\Override] public function m() {}
}
interface I { const A = 1; }
enum E implements I { <error descr="Use the attribute on a permitted target.">#[\Override]</error> case A; }
