<?php
function pick($code, array $allowed) {
    $r = [];
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, $allowed)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\array_search(7, [])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">array_search($code, ['404', '500'])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['  '])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['x' => 'yes'])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($code, ['on', 1])</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('1', [])</weak_warning>;

    /** @var string $color */
    /** @var array $palette */
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($color, $palette)</weak_warning>;

    /** @var string $tone */
    /** @var string[] $tones */
    $r[] = in_array($tone, $tones);

    $r[] = in_array($code, ['draft', "final"]);
    $r[] = array_search($code, array('v1', 'v2'));
    $r[] = in_array($code, $allowed, false);
    $r[] = array_search($code, $allowed, true);
    $r[] = in_array($code);
    return $r;
}
