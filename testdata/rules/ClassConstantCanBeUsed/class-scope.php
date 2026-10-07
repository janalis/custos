<?php

namespace App;

class Base {}

class Child extends Base {
    public function names() {
        $f = function () {
            return <weak_warning descr="Use parent::class instead.">get_parent_class()</weak_warning>;
        };
        return [
            <weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>,
            <weak_warning descr="Use parent::class instead.">\get_parent_class()</weak_warning>,
            $f,
        ];
    }
}

class Root {
    public function names() {
        return [<weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>, get_parent_class()];
    }
}

trait Named {
    public function names() {
        return [<weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>, get_parent_class()];
    }
}

function loose() {
    return [get_called_class(), get_parent_class()];
}

class Host extends Base {
    public function make() {
        function inner() {
            return [get_called_class(), get_parent_class()];
        }
    }
}

$top = [get_called_class(), get_parent_class()];
