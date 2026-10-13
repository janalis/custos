<?php
function collect(array $batches, $c)
{
    $all = [];
    $map = [];
    while ($batch = array_shift($batches)) {
        $map[$batch['id']] = array_merge($map[$batch['id']], $batch);
        if ($batch) {
            $all = array_merge($all, $batch);
        } elseif ($c) {
            $all = array_merge($all, $batch);
        } else {
            $all = array_merge($all, $batch);
        }
        $other = array_merge($all, $batch);
        $all = (array_merge($all, $batch));
        $all = array_merge($all);
        $all = $c->array_merge($all, $batch);
        $fn = function () use ($all, $batch) {
            $all = array_merge($all, $batch);
            return $all;
        };
    }
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        break;
    }
    foreach ($batches as $batch) {
        $all = array_merge($all, $batch);
        return $all;
    }
    $all = array_merge($all, $map);

    for ($n = 0, $total = count($batches); $n < $total; $n++) {}
    for ($n = 0; $n < count($batches) - 1; $n++) {}
    for ($n = 0; $n < (count($batches)); $n++) {}
    for ($n = 0; $n < $c->count(); $n++) {}
    for ($n = 0; $n < size($batches); $n++) {}
    for ($n = 0; ; $n++) {}
    return [$all, $map];
}
