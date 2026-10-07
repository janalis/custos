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
    $now = ob_get_clean();
    

    return [$reader, $arrow, $obj, $now];
}
