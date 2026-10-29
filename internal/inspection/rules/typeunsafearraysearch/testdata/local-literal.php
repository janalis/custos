<?php
function options(array $opts, $k, bool $c)
{
    // A local holding only a literal of non-numeric strings is that literal.
    $known = ['blocking', 'create'];
    foreach ($opts as $key => $v) {
        if (!in_array($key, $known)) {}
    }
    $either = $c ? ['a'] : ['b', 'c'];
    $x1 = in_array($k, $either);

    $grown = ['a', 'b'];
    $grown[] = '1';
    $x2 = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($k, $grown)</weak_warning>;
    $sorted = ['a', 'b'];
    sort($sorted);
    $x3 = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($k, $sorted)</weak_warning>;
    $mixed = $c ? ['a'] : ['1'];
    $x4 = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($k, $mixed)</weak_warning>;
    $x5 = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($k, $opts)</weak_warning>;
    $ref = ['a'];
    $alias = &$ref;
    $x6 = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($k, $ref)</weak_warning>;
    return [$x1, $x2, $x3, $x4, $x5, $x6, $alias];
}
