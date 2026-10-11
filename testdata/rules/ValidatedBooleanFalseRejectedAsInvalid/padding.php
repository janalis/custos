<?php
// ASCII padding is stripped by the filter, so these inputs are valid false spellings.
function padded()
{
    if (<warning descr="Distinguish valid false from invalid boolean input.">filter_var(" off\t\n", FILTER_VALIDATE_BOOLEAN) === false</warning>) {
        return false;
    }
    return true;
}
function scalarOnly()
{
    if (<warning descr="Distinguish valid false from invalid boolean input.">filter_var('No', FILTER_VALIDATE_BOOLEAN, FILTER_REQUIRE_SCALAR) === false</warning>) {
        throw new InvalidArgumentException('not a boolean');
    }
}
