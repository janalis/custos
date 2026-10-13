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
        $c = \is_iterable($src);
        $d = \is_iterable($src);
    }
}

namespace Plain {
    function probe($src)
    {
        // A Plain\is_iterable() declared elsewhere would win over the built-in.
        return \is_iterable($src);
    }
}
