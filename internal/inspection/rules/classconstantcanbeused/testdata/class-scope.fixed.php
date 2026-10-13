<?php

namespace App;

class Base {}

class Child extends Base {
    public function names() {
        $f = function () {
            return parent::class;
        };
        return [
            static::class,
            parent::class,
            $f,
        ];
    }
}

class Root {
    public function names() {
        return [static::class, get_parent_class()];
    }
}

trait Named {
    public function names() {
        return [static::class, get_parent_class()];
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
