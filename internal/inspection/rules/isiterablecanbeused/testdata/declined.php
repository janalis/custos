<?php
function trimAll($prefix)
{
    if (\is_array($prefix) || $prefix instanceof \Traversable) { // don't use is_iterable(), it's slow
        return 1;
    }
    /* a comment above does not count */
    if (<weak_warning descr="Use 'is_iterable($prefix)' instead of the is_array()/instanceof Traversable pair.">is_array($prefix)</weak_warning> || $prefix instanceof Traversable) {
        return 2;
    }
    if (<weak_warning descr="Use 'is_iterable($prefix)' instead of the is_array()/instanceof Traversable pair.">is_array($prefix)</weak_warning> || $prefix instanceof Traversable) { // fast path
        return 3;
    }
    return 0;
}
