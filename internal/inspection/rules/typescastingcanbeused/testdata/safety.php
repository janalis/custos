<?php
function keys($a, $o, $k) {
    $s1 = <weak_warning descr="Use '(string)$a['key']' to make the string conversion explicit.">"$a[key]"</weak_warning>;
    $s2 = <weak_warning descr="Use '(string)$a['01']' to make the string conversion explicit.">"$a[01]"</weak_warning>;
    $s3 = <weak_warning descr="Use '(string)$a[1]' to make the string conversion explicit.">"$a[1]"</weak_warning>;
    $s4 = <weak_warning descr="Use '(string)$a[-1]' to make the string conversion explicit.">"$a[-1]"</weak_warning>;
    $s5 = <weak_warning descr="Use '(string)$a['-0']' to make the string conversion explicit.">"$a[-0]"</weak_warning>;
    $s6 = <weak_warning descr="Use '(string)$a['0x1A']' to make the string conversion explicit.">"$a[0x1A]"</weak_warning>;
    $s7 = <weak_warning descr="Use '(string)$a['99999999999999999999']' to make the string conversion explicit.">"$a[99999999999999999999]"</weak_warning>;
    $s8 = <weak_warning descr="Use '(string)$a[$k]' to make the string conversion explicit.">"$a[$k]"</weak_warning>;
    $s9 = <weak_warning descr="Use '(string)$o->name' to make the string conversion explicit.">"$o->name"</weak_warning>;
    return [$s1, $s2, $s3, $s4, $s5, $s6, $s7, $s8, $s9];
}

function targets(array $row, $i, $flag) {
    settype($row[$i++], 'int');
    settype($row[key()], 'int');
    settype($row[], 'int');
    settype($row[][0], 'int');
    <weak_warning descr="Use '$row['qty'] = (int)$row['qty']' instead (a cast is clearer and faster).">settype($row['qty'], 'int')</weak_warning>;
    <weak_warning descr="Use '$this->qty = (int)$this->qty' instead (a cast is clearer and faster).">settype($this->qty, 'int')</weak_warning>;
    <weak_warning descr="Use '$total = (float)$total' instead (a cast is clearer and faster).">settype($total, 'float')</weak_warning>;
    $sum = $row['sum'];
    <weak_warning descr="Use '$sum = (float)$sum' instead (a cast is clearer and faster).">settype($sum, 'float')</weak_warning>;
    if ($flag) {
        <weak_warning descr="Use '$sum = (string)$sum' instead (a cast is clearer and faster).">settype($sum, 'string')</weak_warning>;
        $late = 1;
    }
    if ($flag) <weak_warning descr="Use '$sum = (int)$sum' instead (a cast is clearer and faster).">settype($sum, 'int')</weak_warning>;
    <weak_warning descr="Use '$late = (int)$late' instead (a cast is clearer and faster).">settype($late, 'int')</weak_warning>;
    $gone = 1;
    unset($row['x'], $gone);
    <weak_warning descr="Use '$gone = (int)$gone' instead (a cast is clearer and faster).">settype($gone, 'int')</weak_warning>;
    $outer = 1;
    $fn = function ($inner) {
        <weak_warning descr="Use '$outer = (int)$outer' instead (a cast is clearer and faster).">settype($outer, 'int')</weak_warning>;
        <weak_warning descr="Use '$inner = (int)$inner' instead (a cast is clearer and faster).">settype($inner, 'int')</weak_warning>;
    };
    switch ($i) {
        case 1:
            $picked = $i;
            <weak_warning descr="Use '$picked = (bool)$picked' instead (a cast is clearer and faster).">settype($picked, 'bool')</weak_warning>;
    }
    try {
        risky();
    } catch (\Exception $e) {
        <weak_warning descr="Use '$sum = (int)$sum' instead (a cast is clearer and faster).">settype($sum, 'int')</weak_warning>;
    }
    if ($i) {
        risky();
    } elseif ($flag) {
        <weak_warning descr="Use '$sum = (int)$sum' instead (a cast is clearer and faster).">settype($sum, 'int')</weak_warning>;
    } else {
        <weak_warning descr="Use '$sum = (int)$sum' instead (a cast is clearer and faster).">settype($sum, 'int')</weak_warning>;
    }
    return $fn;
}

function callees($h, $cls, $o) {
    $a = <weak_warning descr="Use '((string)$h)' to make the string conversion explicit.">"$h"</weak_warning>();
    $b = <weak_warning descr="Use '((string)$cls)' to make the string conversion explicit.">"$cls"</weak_warning>::make();
    $c = <weak_warning descr="Use '((string)$h)' instead (a cast is clearer and faster).">strval($h)</weak_warning>();
    $d = <weak_warning descr="Use '((string)$cls)' to make the string conversion explicit.">"$cls"</weak_warning>::LIMIT;
    $e = <weak_warning descr="Use '((string)$cls)' to make the string conversion explicit.">"$cls"</weak_warning>::$count;
    $f = <weak_warning descr="Use '((string)$o)' instead of calling __toString() directly.">$o->__toString()</weak_warning>();
    $g = $o->__toString(...);
    $h = $h(<weak_warning descr="Use '(string)$cls' to make the string conversion explicit.">"$cls"</weak_warning>);
    $i = Registry::get(<weak_warning descr="Use '(string)$cls' to make the string conversion explicit.">"$cls"</weak_warning>);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}

$top = '5';
<weak_warning descr="Use '$top = (int)$top' instead (a cast is clearer and faster).">settype($top, 'int')</weak_warning>;
<weak_warning descr="Use '$missing = (int)$missing' instead (a cast is clearer and faster).">settype($missing, 'int')</weak_warning>;
