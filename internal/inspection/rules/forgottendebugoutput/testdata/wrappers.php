<?php
namespace Acme;

class Tracer {
    public static function dump($v) {
        var_dump($v);
    }
    public static function show($v) {
        <error descr="Debug output call; remove it if it was left over from debugging.">print_r($v)</error>;
    }
}

class Other {
    public function dump($v) {
        var_dump($v);
    }
    public function dd(...$vars) {
        dump(...$vars);
        exit(1);
    }
    public function render($v) {
        <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($v)</error>;
    }
}

function run($x) {
    <error descr="Debug output call; remove it if it was left over from debugging.">Tracer::dump($x)</error>;
}
