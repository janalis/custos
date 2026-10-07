<?php

namespace Kernel;

function get_called_class(): string { return 'x'; }

class Base {}

class Node extends Base {
    public function names() {
        return [
            // Kernel\get_called_class() is a user function.
            get_called_class(),
            static::class,
            parent::class,
        ];
    }
}
