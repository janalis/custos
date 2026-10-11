<?php
function fetchWithRetry(Client $client, string $url) {
    $failure = null;
    for ($attempt = 0; $attempt < 3; $attempt++) {
        try {
            return $client->get($url);
        } catch (\RuntimeException $failure) {      // E2b: read after the loop
            usleep(100);
        }
    }
    throw new \LogicException('gave up', 0, $failure);
}

function cleanup($handle) {
    try {
        flush_buffers($handle);
    } catch (\Exception $pending) {               // E2b: used in finally
    } finally {
        fclose($handle);
        if (isset($pending)) { report($pending); }
    }
    try {
        rewind_all($handle);
    } catch (\Exception <weak_warning descr="Caught exception is silently discarded; at least log it.">$oops</weak_warning>) {
    }
    try {
        close_all($handle);
    } catch (\Exception $oops) {                  // same name rebound: does not save the first one
        report($oops);
        $cb = fn() => $oops;
    }
    $later = function () { return $oops; };       // separate scope: ignored
}

function captured() {
    try {
        work();
    } catch (\Exception $e) {                     // E2b: captured by a later closure use list
    }
    return function () use ($e) { return $e; };
}

function arrow() {
    try {
        work();
    } catch (\Exception $e) {                     // E2b: captured by a later arrow function
        cleanup_all();
    }
    return fn() => $e;
}

function before() {
    $x = $e ?? null;
    try {
        work();
    } catch (\Exception <weak_warning descr="Caught exception is dropped; log it or pass it on as the previous exception.">$e</weak_warning>) {
        cleanup_all();
    }
    class Inner { function m() { return $e; } }   // separate scope: ignored
}

try {
    work();
} catch (\Exception $top) {                       // E2b: top-level code
}
echo $top;

function nextIteration(callable $op) {
    $attempt = 0;
    while (true) {
        if (isset($failure)) {
            report($failure);
            if (++$attempt >= 2) { return null; }
        }
        try {
            return $op();
        } catch (\RuntimeException $failure) {     // E2c: read at the top of the next iteration
        }
    }
}

function loopHeader(callable $op) {
    $last = null;
    do {
        foreach ([1, 2] as $step) {
            try {
                $op($step);
            } catch (\RuntimeException $last) {    // E2c: read in the outer loop condition
                usleep(100);
            }
        }
    } while ($last !== null && retry($last));
    for ($i = 0; $i < 2; $i++) {
        try {
            $op($i);
        } catch (\LogicException <weak_warning descr="Caught exception is silently discarded; at least log it.">$gone</weak_warning>) {
        }
    }
}

function beforeLoop(callable $op) {
    $seen = $early ?? null;
    while (more()) {
        $cb = function () { return $early; };     // separate scope: ignored
        try {
            $op();
        } catch (\Exception <weak_warning descr="Caught exception is dropped; log it or pass it on as the previous exception.">$early</weak_warning>) {
            usleep(100);
        }
    }
}

while (more()) {
    $run = function () {
        try {
            work();
        } catch (\Exception <weak_warning descr="Caught exception is silently discarded; at least log it.">$outer</weak_warning>) {
        }
    };
    echo $outer ?? '';                            // the loop is outside the closure: ignored
}
