<?php
class Crate {
    public int $count = 0;
    public ?string $label = null;
}

function probes(Crate $crate, ?Crate $maybe, int $num, $any, array $list) {
    $a = <weak_warning descr="Simplify to '$crate->count ?? 7' using the null coalescing operator.">!empty($crate) ? $crate->count : 7</weak_warning>;
    $b = <weak_warning descr="Simplify to '$maybe->count ?? 7' using the null coalescing operator.">$maybe ? $maybe->count : 7</weak_warning>;
    $c = <weak_warning descr="Simplify to '$any->label ?? null' using the null coalescing operator.">!empty($any) ? $any->label : null</weak_warning>;
    $d = <weak_warning descr="Simplify to '$maybe->label ?? null' using the null coalescing operator.">$maybe ? $maybe->label : null</weak_warning>;

    $e = !empty($crate) ? $crate->label : 'none';   // nullable property, non-null fallback
    $f = !empty($any) ? $any->count : 7;            // probe type unknown
    $g = $num ? $num->count : 7;                    // probe is a scalar
    $h = !empty($list[0]) ? $list[0]->count : 7;    // probe may be a scalar
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}

function previous($in) {
    $v = 'plain';
    <weak_warning descr="Simplify to '$v = $in['v'] ?? 'plain'' using the null coalescing operator.">if</weak_warning> (isset($in['v'])) {
        $v = $in['v'];
    }
    $w = compute();
    if (isset($in['w'])) {
        $w = $in['w'];
    }
    $x = new Crate();
    if (isset($in['x'])) {
        $x = $in['x'];
    }
    return [$v, $w, $x];
}
