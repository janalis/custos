<?php
namespace App {
    final class Invoice {}

    function is_a($value, $class) { return false; }
    function class_parents($value) { return []; }

    function local(Invoice $doc) {
        return [
            is_a($doc, 'App\\Invoice'),
            \App\is_a($doc, 'App\\Invoice'),
            in_array('App\\Invoice', class_parents($doc)),
            $doc instanceof \App\Invoice,
        ];
    }
}

namespace Other {
    use function App\is_a;

    function imported(\App\Invoice $doc) {
        return [
            is_a($doc, 'App\\Invoice'),
            $doc instanceof \App\Invoice,
        ];
    }
}
