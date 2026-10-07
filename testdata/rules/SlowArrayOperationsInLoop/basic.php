<?php
function collect(array $batches, $repo)
{
    $all = [];
    $map = [];
    foreach ($batches as $batch) {
        $all = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge($all, $batch)</error>;
        $map = <error descr="'array_replace(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">\array_replace($batch, $map)</error>;
        try {
            $repo->rows = <error descr="'array_merge_recursive(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge_recursive($repo->rows, $batch)</error>;
        } finally {
        }
    }
    while ($batch = array_shift($batches))
        $all += <error descr="'array_replace_recursive(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_replace_recursive($all, $batch)</error>;
    do {
        self::$cache = <error descr="'array_merge(...)' inside a loop re-copies the accumulator each time; merge once after the loop.">array_merge(self::$cache, $batches)</error>;
    } while (false);

    for ($n = 0; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">$n < count($batches)</error>; $n++) {}
    for ($n = 0, $m = 1; <error descr="'strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">strlen($repo->name) >= $n</error>; $n++) {}
    for ($repo->at = 0; <error descr="'mb_strlen(...)' is re-evaluated on every iteration; compute it once before the loop.">$repo->at <= mb_strlen($repo->name)</error>; $repo->at++) {}
    for (; <error descr="'count(...)' is re-evaluated on every iteration; compute it once before the loop.">count($batches) > $n</error>; $n++) {}
    for ($n = 0; <error descr="'sizeof(...)' is re-evaluated on every iteration; compute it once before the loop.">$n < \sizeof($batches)</error>; $n++) {}
    return [$all, $map];
}
