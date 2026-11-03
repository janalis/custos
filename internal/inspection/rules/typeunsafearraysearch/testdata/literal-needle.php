<?php
function roles(array $list, $user) {
    $r = [];
    $r[] = in_array('admin', $list);
    $r[] = array_search(" editor ", $list);
    $r[] = in_array(\DateTime::class, $list);
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('42', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array(' ', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array("$user", $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array(7, $list)</weak_warning>;
    return $r;
}

function loops(array $list, $other) {
    $r = [];
    foreach (['sValue1', 'sValue2'] as $k) {
        $r[] = in_array($k, $list);
        echo $k;
    }
    $role = 'admin';
    $r[] = array_search($role, $list);
    foreach (['a', '1'] as $n) {
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($n, $list)</weak_warning>;
    }
    foreach (['a', 'b'] as &$ref) {
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($ref, $list)</weak_warning>;
    }
    foreach (['a', 'b'] as $w) {
        $w = $other;
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($w, $list)</weak_warning>;
    }
    foreach (['a', 'b'] as $s) {
        settype($s, 'int');
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($s, $list)</weak_warning>;
    }
    foreach (['a', 'b'] as $e) {
        $e[0] = '1';
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($e, $list)</weak_warning>;
    }
    foreach (['a', 'b'] as $o) {
    }
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($o, $list)</weak_warning>;
    $p = $other;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($p, $list)</weak_warning>;
    return $r;
}

function reused(array $list, array $more) {
    $r = [];
    $t = 'x';
    foreach ($more as $t) {
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, $more)</weak_warning>;
    }
    foreach (['a', 'b'] as $t) {
        $r[] = in_array($t, $list);
    }
    $alias = &$u;
    foreach (['a', 'b'] as $u) {
        $alias = 1;
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($u, $list)</weak_warning>;
    }
    $refs = [&$v];
    foreach (['a', 'b'] as $v) {
        $refs[0] = 1;
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($v, $list)</weak_warning>;
    }
    $plain = [$v];
    $copy = $v;
    $v .= 'x';
    static $g;
    foreach (['a', 'b'] as $g) {
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($g, $list)</weak_warning>;
    }
    return $r;
}

function through(array $info, bool $old) {
    $r = [];
    $keys = ['redis_version', 'arch_bits'];
    foreach ($keys as $k) {
        $r[] = in_array($k, array_keys($info));
    }
    $more = ['a', 'b'];
    if ($old) {
        array_push($more, '1');
    }
    foreach ($more as $m) {
        $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($m, $info)</weak_warning>;
    }
    return $r;
}

function concatenated(array $list, $op, $id, $n) {
    $r = [];
    $r[] = in_array($op . '_any', $list);
    $r[] = in_array('field_' . $id, $list);
    $r[] = in_array(('k' . $id) . $n . '.', $list);
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($op . '1', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($op . "\n", $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('-' . $id, $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($op . '', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('' . $op, $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($op . $id, $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($op . "x$id", $list)</weak_warning>;
    return $r;
}
