<?php
function f() {
    goto later;
    later:
    $cleanup = function () {
        finish:
        return true;
        goto finish;
    };
}
class A {
    public function m() {
        loop:
        if (rand()) { while (true) { goto loop; } }
    }
}

start:
echo "top level";

// Top-level labels are not inspected (E1).
top:
echo 1;
