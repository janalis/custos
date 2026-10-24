<?php
function collect(string $dir, string $p)
{
    // An element written with a glob() result does not make the array one.
    $files = [];
    foreach (glob($dir . '/*.yml') ?: [] as $file) {
        $files[] = $file;
    }
    $files = array_merge($files, ['extra']);
    $byName = [];
    $byName['x'] = glob($p);
    echo count($byName);
    $list = glob($p);
    // Writing into a failed result still leaves the failure unhandled.
    $list['extra'] = 'x';
    echo count(<warning descr="Handle glob failure before using its result as an array.">$list</warning>);

    // A path that built an array adds no producer to the others.
    $found = glob($p);
    if ($dir === 'b') {
        $found = glob($dir);
    }
    if ($dir === 'c') {
        $found = [];
        $found[] = $p;
    }
    echo count(<warning descr="Handle glob failure before using its result as an array.">$found</warning>);
}
