<?php
namespace Acme;

use function var_dump;

class Tracer { public static function dump($v) {} }
class Probe  { public function dump($v) {} }
class Quiet  { public function dump($v) {} }
class SubProbe extends Probe {}
class OwnProbe extends Probe { public function dump($v) {} }

function audit_dump($v) {
    print_r($v);
}

function work($order, Probe $probe, Quiet $quiet, SubProbe $sub, OwnProbe $own, ?Probe $maybe) {
    <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($order, $probe)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">\print_r($order)</error>;
    $text = print_r($order, true);
    echo var_export($order, true);
    <error descr="Debug output call; remove it if it was left over from debugging.">var_export($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">error_log('order seen', 3, '/var/log/app.log')</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">debug_print_backtrace()</error>;
    phpinfo(INFO_GENERAL);
    <error descr="Debug output call; remove it if it was left over from debugging.">phpinfo()</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">audit_dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">Tracer::dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">$probe->dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">$sub->dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">$maybe?->dump($order)</error>;
    $own->dump($order);
    $quiet->dump($order);
    <error descr="Debug output call; remove it if it was left over from debugging.">$probe->Dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">VAR_DUMP($order)</error>;

    ob_start();
    var_dump($order);
    ob_start();
    @print_r($order);
    ob_start();
    echo <error descr="Debug output call; remove it if it was left over from debugging.">print_r($order)</error>;
    ob_start();
    $copy = $order;
    <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($copy)</error>;

    $f = function () use ($order) { <error descr="Debug output call; remove it if it was left over from debugging.">var_dump($order)</error>; };
}

function dd($v) {
    var_dump($v);
}
