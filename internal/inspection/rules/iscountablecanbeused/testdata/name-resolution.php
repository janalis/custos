<?php

namespace Shop {
    use Countable as Sized;

    function is_array($v): bool { return false; }

    function probe($box)
    {
        // Shop\is_array() is not the global function.
        $a = is_array($box) || $box instanceof \Countable;
        // Countable resolves to Shop\Countable here.
        $b = \is_array($box) || $box instanceof Countable;
        $c = <weak_warning descr="Use 'is_countable($box)' instead of the is_array()/instanceof Countable pair.">\IS_ARRAY($box)</weak_warning> || $box instanceof \countable;
        $d = <weak_warning descr="Use 'is_countable($box)' instead of the is_array()/instanceof Countable pair.">\is_array($box)</weak_warning> || $box instanceof Sized;
    }
}
