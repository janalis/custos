<?php
namespace App;
function helper($cb) {}
class Svc {
    public function run() {
        helper(function () { return 1; });
        unknown_fn(function () { return 1; });
        $f = function () { return 2; };
        $f->bind($f, null);
        $g = function () { return 3; };
        \Closure::bind($g, $this);
    }
}
