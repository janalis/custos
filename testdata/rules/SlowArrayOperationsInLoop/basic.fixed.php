<?php
function collect(array $batches, $repo)
{
    $all = [];
    $map = [];
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        $map = \array_replace($batch, $map);
        try {
            $repo->rows = array_merge_recursive($repo->rows, $batch);
        } finally {
        }
    }
    while ($batch = array_shift($batches))
        $all += array_replace_recursive($all, $batch);
    do {
        self::$cache = array_merge(self::$cache, $batches);
    } while (false);

    for ($n = 0, $nMax = count($batches); $n < $nMax; $n++) {}
    for ($n = 0, $m = 1, $nMax = strlen($repo->name); $nMax >= $n; $n++) {}
    for ($repo->at = 0, $loopsMax = mb_strlen($repo->name); $repo->at <= $loopsMax; $repo->at++) {}
    for ($nMax = count($batches); $nMax > $n; $n++) {}
    for ($n = 0, $nMax = \sizeof($batches); $n < $nMax; $n++) {}
    return [$all, $map];
}
