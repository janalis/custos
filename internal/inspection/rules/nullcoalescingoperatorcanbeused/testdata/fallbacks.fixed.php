<?php
class Crate {
    public int $count = 0;
    public ?string $label = null;
}

function probes(Crate $crate, ?Crate $maybe, int $num, $any, array $list) {
    $a = $crate->count ?? 7;
    $b = $maybe->count ?? 7;
    $c = $any->label ?? null;
    $d = $maybe->label ?? null;

    $e = !empty($crate) ? $crate->label : 'none';   // nullable property, non-null fallback
    $f = !empty($any) ? $any->count : 7;            // probe type unknown
    $g = $num ? $num->count : 7;                    // probe is a scalar
    $h = !empty($list[0]) ? $list[0]->count : 7;    // probe may be a scalar
    return [$a, $b, $c, $d, $e, $f, $g, $h];
}

function previous($in) {
    $v = $in['v'] ?? 'plain';
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
