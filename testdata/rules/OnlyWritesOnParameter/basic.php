<?php
function withInclude($n, $path)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$n</weak_warning> -= 1;
    require $path;
}

function reachability($a, $b)
{
    if ($b) {
        return;
    }
    <weak_warning descr="Value is only written here and never read; the write is lost.">$a[]</weak_warning> = 1;
    return;
    echo $a;
}

function chained(array $list, ...$rest)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$list['a']['b']</weak_warning> = 1;
    <weak_warning descr="Value is only written here and never read; the write is lost.">$rest[]</weak_warning> = 2;
    if (false === (<weak_warning descr="Variable is never used.">$found</weak_warning> = array_search(1, [1]))) {
        return null;
    }
}

function viaCustos(array $out)
{
    /** @custos-ignore OnlyWritesOnParameter */
    $out[] = 1;
}

function destructured($left, $right, $item, array $pairs)
{
    [<weak_warning descr="Value is only written here and never read; the write is lost.">$left</weak_warning>, 'k' => [<weak_warning descr="Value is only written here and never read; the write is lost.">$right</weak_warning>]] = $pairs;
    foreach ($pairs as [<weak_warning descr="Value is only written here and never read; the write is lost.">$item</weak_warning>]) {
    }
}
