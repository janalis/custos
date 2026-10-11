<?php
function assigned() {
    $g = (function () { yield 1; return 2; })();
    $r = <warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>;
    return $r;
}
function returned() {
    $g = (function () { yield 1; return 2; })();
    return <warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>;
}
function echoed() {
    $g = (function () { yield 1; return 2; })();
    echo <warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>;
}
function lowercase() {
    $g = (function () { yield 1; return 2; })();
    $r = (int) <warning descr="Finish the generator before reading its return value.">$g->getreturn()</warning>;
    return $r;
}
function passed() {
    $g = (function () { yield 1; return 2; })();
    var_dump(<warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>);
}
