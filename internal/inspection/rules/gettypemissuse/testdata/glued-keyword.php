<?php
// The literal directly follows a keyword: the predicate needs a separator.
function glued($x)
{
    echo<warning descr="Use 'is_int($x)' instead.">'integer'==gettype($x)</warning>;
    echo<warning descr="Use '!is_int($x)' instead.">'integer'!=gettype($x)</warning>;
    return<warning descr="Use 'is_string($x)' instead.">'string'===gettype($x)</warning>;
}
