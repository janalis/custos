<?php

namespace App;

use Lib\Kinds;

function kinds($t)
{
    // Constants declared in another file whose value is a plain string.
    in_array($t, [Kinds::PLAIN, Kinds::QUOTED]);

    // Numeric strings and initializers that are not one plain literal.
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::PLAIN, Kinds::NUMERIC])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::ESCAPED])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::JOINED])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::VARIABLE])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::MISSING])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [static::PLAIN])</weak_warning>;
    <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array($t, [Kinds::{'PLAIN'}])</weak_warning>;
}
