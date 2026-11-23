<?php
// PHP 8.6 adds class constants and enum cases.
class P { public $x; const X = 1; public function m() {} }
class C extends P {
    #[\Override] public $x;
    #[\Override] const X = 2;
    #[\Override] public function m() {}
}
interface I { const A = 1; }
enum E implements I { #[\Override] case A; }
<error descr="Use the attribute on a permitted target.">#[\Override]</error> function f() {}
