<?php
array_rand(['oak', 'elm'], 1);

function pickSeen(array $words)
{
    $seen = [];
    foreach ($words as $word) {
        @$seen[$word]++;
    }
    return array_rand($seen);
}

function pickSome(array $pool, array $wanted)
{
    $n = 0;
    foreach ($wanted as $w) {
        if ($w) {
            $n++;
        }
    }
    return array_rand($pool, $n);
}

function pickSlot(array $rows)
{
    $slots = [];
    foreach ($rows as $slots['last']) {
    }
    return array_rand($slots);
}

function pickShared()
{
    $shared = [];
    $fill = function () use (&$shared) {
        $shared[] = 'ash';
    };
    $fill();
    return array_rand($shared);
}

function pickBoxed()
{
    $boxed = [];
    $box = [&$boxed];
    $box[0][] = 'yew';
    return array_rand($boxed);
}

function pickExtracted(array $vars)
{
    $picked = [];
    extract($vars);
    return array_rand($picked);
}

function pickCopied(array $words)
{
    $seen = [];
    foreach ($words as $word) {
        @$seen[$word]++;
    }
    $pool = $seen;
    return array_rand($pool);
}

function pickSlotted(array $pool)
{
    $n = 0;
    foreach ([&$n] as &$slot) {
        $slot = 1;
    }
    return array_rand($pool, $n);
}
