<?php
class Widget {
    public function name() {
        $f = fn() => <weak_warning descr="Use the __CLASS__ constant instead of this call.">get_class()</weak_warning>;
        return <weak_warning descr="Use the __CLASS__ constant instead of this call.">get_class()</weak_warning>;
    }
}

trait Named {
    public function name() {
        return <weak_warning descr="Use the __CLASS__ constant instead of this call.">get_class()</weak_warning>;
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
