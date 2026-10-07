<?php
/** @param int[] $ids */
function pick(array $rows, iterable $feed, $ids, callable $cb, float ...$nums) {
    $ok = is_array($rows) && is_array($feed) && is_callable($cb) && is_object($cb);
    $bad = <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_string($rows)</warning>
        || <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_bool($feed)</warning>
        || !<warning descr="This check is always true for the declared parameter type; is the parameter being reused?">is_float($ids)</warning>
        || <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">\is_float($nums)</warning>;
    return [$ok, $bad];
}

interface Shape {}
class Circle implements Shape {}
class Ring extends Circle {}
class Label {}

function redraw(Shape $s, Ring $r = null, int $n = 0, string $txt = '', array $list = []) {
    $s = $s ?? new Circle();
    $r = $r ?? new Circle();
    $r = <warning descr="Assigning a value of type \Label does not match the parameter's declared type.">new Label()</warning>;
    $n = 2 * $n + 1;
    $n = <warning descr="Assigning a value of type float does not match the parameter's declared type.">$n / 2.5</warning>;
    $txt = str_replace('a', 'b', $txt);
    $txt = substr($txt, 1);
    $list = explode(',', $txt);
    $list = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'none'</warning>;
    $s = <warning descr="Assigning a value of type null does not match the parameter's declared type.">null</warning>;
    $txt .= 5;
    $txt = <warning descr="Assigning a value of type array does not match the parameter's declared type.">$_GET['q'] ?? ''</warning>;
    $n = $_SERVER['REQUEST_TIME'];
}

class Builder {
    public function make(Shape $shape, int $count) {
        $shape = <warning descr="Assigning a value of type \Builder does not match the parameter's declared type.">$this</warning>;
        if (<warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_string($count)</warning>) {
            return;
        }
        $f = function () use ($count) { return is_string($count); };
    }
}
