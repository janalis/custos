<?php
function braceless($x)
{
    if ($x) return;
    if ($x):
        return;
    endif;
    echo 1;
}

function notLast()
{
    return;
    echo 1;
}

function caught($y)
{
    try {
        <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $y = compute();</weak_warning>
    } catch (\Exception $e) {
        return null;
    }
}

function closingTag()
{
    $z = 0;
    <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $z = 5 ?>
</weak_warning><?php
}
