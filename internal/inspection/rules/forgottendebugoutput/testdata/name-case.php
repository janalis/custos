<?php
namespace Acme;

class TRACER { public static function DUMP($v) {} }
class Logger { public function dump($v) {} }

function AUDIT_DUMP($v) {
    // listed wrapper declared in another case: not reported
    VAR_DUMP($v);
}

function inspect($order, Logger $log) {
    <error descr="Debug output call; remove it if it was left over from debugging.">Var_Dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">\PRINT_R($order)</error>;
    $text = Print_R($order, TRUE);
    <error descr="Debug output call; remove it if it was left over from debugging.">tracer::Dump($order)</error>;
    <error descr="Debug output call; remove it if it was left over from debugging.">Audit_Dump($order)</error>;
    $log->DUMP($order);
    OB_START();
    var_dump($order);
}
