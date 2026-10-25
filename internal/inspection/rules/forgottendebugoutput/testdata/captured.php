<?php
class Probe { public function dump($v) {} public static function show($v) {} }

function capture($order, Probe $probe) {
    ob_start();
    var_dump($order);
    print_r($order);
    $returned = var_export($order);
    if ($order) {
        echo print_r($order);
    }
    $printed = ob_get_clean();

    ob_start(fn ($b) => strtoupper($b));
    $noted = 1;
    var_dump($noted);
    ob_end_flush();

    ob_start();
    ob_end_clean();
    <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($order)</error>;
    $late = ob_get_contents();

    ob_start();
    $f = function () use ($order) { <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($order)</error>; };
    <error descr="Debug output call; remove it if it was left over from debugging.">print_r($order)</error>;
    $g = function () { ob_get_clean(); };

    $dumper = var_dump(...);
    $method = $probe->dump(...);
    $static = Probe::show(...);
    return [$printed, $returned, $late, $dumper, $method, $static];
}
