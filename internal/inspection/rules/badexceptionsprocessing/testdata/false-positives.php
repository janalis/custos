<?php
function f() {
    try {
        a();
        b();
        ;
        c();
    } catch (\Exception $e) {
        echo "failed: $e";
    } catch (\Throwable $t) {
        $fn = fn() => $t;
    }
}
