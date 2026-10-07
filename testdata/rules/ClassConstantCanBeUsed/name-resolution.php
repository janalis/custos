<?php

namespace Kernel;

function get_called_class(): string { return 'x'; }

class Base {}

class Node extends Base {
    public function names() {
        return [
            // Kernel\get_called_class() is a user function.
            get_called_class(),
            <weak_warning descr="Use static::class instead.">\GET_CALLED_CLASS()</weak_warning>,
            <weak_warning descr="Use parent::class instead.">Get_Parent_Class()</weak_warning>,
        ];
    }
}
