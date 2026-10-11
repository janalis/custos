<?php
// The literal directly follows a keyword: the predicate needs a separator.
function glued($x)
{
    echo is_int($x);
    echo!is_int($x);
    return is_string($x);
}
