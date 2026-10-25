<?php
function locked($lock) {
    $lock->acquire();
    try {
        prepare();
        run();
        report();
        flush();
    } finally {
        $lock->release();
    }

    <weak_warning descr="Too many statements in this try block; extract some of them so the failing call is obvious.">try</weak_warning> {
        prepare();
        run();
        report();
        flush();
    } catch (\RuntimeException $e) {
        throw $e;
    } finally {
        $lock->release();
    }
}
