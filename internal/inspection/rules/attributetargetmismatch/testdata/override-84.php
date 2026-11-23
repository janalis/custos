<?php
// Before PHP 8.5, PHP's own Override targets methods only.
class P { public $x; const X = 1; public function m() {} }
class C extends P {
    <error descr="Use the attribute on a permitted target.">#[\Override]</error> public $x;
    <error descr="Use the attribute on a permitted target.">#[\Override]</error> const X = 2;
    #[\Override] public function m() {}
}
// PHP 8.4's Deprecated targets functions, methods and class constants.
<error descr="Use the attribute on a permitted target.">#[\Deprecated]</error> class Old {}
#[\Deprecated] function oldHelper() {}
class Legacy { #[\Deprecated] const OLD = 1; #[\Deprecated] public function run() {} }
