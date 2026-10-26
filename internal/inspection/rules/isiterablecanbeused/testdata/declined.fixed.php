<?php
function trimAll($prefix)
{
    if (\is_array($prefix) || $prefix instanceof \Traversable) { // don't use is_iterable(), it's slow
        return 1;
    }
    /* a comment above does not count */
    if (is_iterable($prefix)) {
        return 2;
    }
    if (is_iterable($prefix)) { // fast path
        return 3;
    }
    return 0;
}
