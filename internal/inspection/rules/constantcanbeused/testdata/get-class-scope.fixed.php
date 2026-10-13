<?php
class Widget {
    public function name() {
        $f = fn() => __CLASS__;
        return __CLASS__;
    }
}

trait Named {
    public function name() {
        return __CLASS__;
    }
}

function helper() {
    return get_class();
}

class Host {
    public function make() {
        function inner() {
            return get_class();
        }
    }
}

$top = get_class();
