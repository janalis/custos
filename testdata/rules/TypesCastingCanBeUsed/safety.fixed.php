<?php
function keys($a, $o, $k) {
    $s1 = (string)$a['key'];
    $s2 = (string)$a['01'];
    $s3 = (string)$a[1];
    $s4 = (string)$a[-1];
    $s5 = (string)$a['-0'];
    $s6 = (string)$a['0x1A'];
    $s7 = (string)$a['99999999999999999999'];
    $s8 = (string)$a[$k];
    $s9 = (string)$o->name;
    return [$s1, $s2, $s3, $s4, $s5, $s6, $s7, $s8, $s9];
}

function targets(array $row, $i, $flag) {
    settype($row[$i++], 'int');
    settype($row[key()], 'int');
    settype($row[], 'int');
    settype($row[][0], 'int');
    settype($row['qty'], 'int');
    settype($this->qty, 'int');
    settype($total, 'float');
    $sum = $row['sum'];
    $sum = (float)$sum;
    if ($flag) {
        $sum = (string)$sum;
        $late = 1;
    }
    if ($flag) $sum = (int)$sum;
    settype($late, 'int');
    $gone = 1;
    unset($row['x'], $gone);
    settype($gone, 'int');
    $outer = 1;
    $fn = function ($inner) {
        settype($outer, 'int');
        $inner = (int)$inner;
    };
    switch ($i) {
        case 1:
            $picked = $i;
            $picked = (bool)$picked;
    }
    try {
        risky();
    } catch (\Exception $e) {
        $sum = (int)$sum;
    }
    if ($i) {
        risky();
    } elseif ($flag) {
        $sum = (int)$sum;
    } else {
        $sum = (int)$sum;
    }
    return $fn;
}

function callees($h, $cls, $o) {
    $a = ((string)$h)();
    $b = ((string)$cls)::make();
    $c = ((string)$h)();
    $d = ((string)$cls)::LIMIT;
    $e = ((string)$cls)::$count;
    $f = ((string)$o)();
    $g = $o->__toString(...);
    $h = $h((string)$cls);
    $i = Registry::get((string)$cls);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i];
}

$top = '5';
$top = (int)$top;
settype($missing, 'int');
