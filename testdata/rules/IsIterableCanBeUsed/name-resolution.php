<?php

namespace Feed {
    use Traversable as Walkable;

    function is_array($v): bool { return false; }

    function probe($src)
    {
        // Feed\is_array() is not the global function.
        $a = is_array($src) || $src instanceof \Traversable;
        // Traversable resolves to Feed\Traversable here.
        $b = \is_array($src) || $src instanceof Traversable;
        $c = <weak_warning descr="Use 'is_iterable($src)' instead of the is_array()/instanceof Traversable pair.">\Is_Array($src)</weak_warning> || $src instanceof \TRAVERSABLE;
        $d = <weak_warning descr="Use 'is_iterable($src)' instead of the is_array()/instanceof Traversable pair.">\is_array($src)</weak_warning> || $src instanceof Walkable;
    }
}
