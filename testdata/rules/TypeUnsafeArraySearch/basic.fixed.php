<?php
function pick($code, array $allowed) {
    $r = [];
    $r[] = in_array($code, $allowed, true);
    $r[] = \array_search(7, [], true);
    $r[] = array_search($code, ['404', '500'], true);
    $r[] = in_array($code, ['  '], true);
    $r[] = in_array($code, ['x' => 'yes'], true);
    $r[] = in_array($code, ['on', 1], true);
    $r[] = in_array('1', [], true);

    /** @var string $color */
    /** @var array $palette */
    $r[] = in_array($color, $palette, true);

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
