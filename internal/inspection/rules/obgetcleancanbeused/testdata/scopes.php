<?php
function deferred() {
    ob_start();
    $reader = function () { return ob_get_contents(); };
    ob_end_clean();

    ob_start();
    $arrow = fn() => ob_get_contents();
    ob_end_clean();

    ob_start();
    $obj = new class { public function read() { return ob_get_contents(); } };
    ob_end_clean();

    ob_start();
    send(ob_get_contents(), ob_get_contents());
    ob_end_clean();

    ob_start();
    $now = <warning descr="Use ob_get_clean() instead of ob_get_contents() + ob_end_clean().">ob_get_contents()</warning>;
    ob_end_clean();

    return [$reader, $arrow, $obj, $now];
}
