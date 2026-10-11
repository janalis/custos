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
        return compute();
    } catch (\Exception $e) {
        return null;
    }
}

function closingTag()
{
    $z = 0;
    return 5 ?>
<?php
}

function nearMisses(array $xs, $o)
{
    foreach ($xs as $n) {
        $copy = &$o->p;
        $pair = [&$copy, $n];
        try {
            try {
                return 3;
            } finally {
                echo $copy;
            }
        } finally {
            echo $pair[0];
        }
    }
}
