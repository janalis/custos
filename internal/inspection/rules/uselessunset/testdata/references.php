<?php
// Unsetting a name bound to a reference detaches it: later writes stay local.
function detach(array &$rows)
{
    $copy = $rows;
    unset($rows);
    $rows = array_reverse($copy);
    return $rows;
}

function detachInLoop(array &$acc, array $items)
{
    foreach ($items as $item) {
        $acc = $item;
        if ($item === null) {
            unset($acc);
        }
    }
}

function aliasTarget($p, array &$store)
{
    $p = &$store['slot'];
    unset($p);
    $p = 1;
}

function aliasSource($p)
{
    $keep = &$p;
    unset($p);
    $p = 2;
    return $keep;
}

function loopAlias($row, array $rows)
{
    foreach ($rows as &$row) {
        $row = trim($row);
    }
    unset($row);
    $row = null;
}

function captured($count)
{
    $bump = function () use (&$count, $bump) { $count++; };
    unset($count);
    $count = 0;
    return $bump;
}

// Nothing touches the name afterwards, or only detaches it again.
function untouched(array &$queue, &$again, &$g, &$s)
{
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($queue);</weak_warning>
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($again);</weak_warning>
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($g);</weak_warning>
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($s);</weak_warning>
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($again);</weak_warning>
    global $g;
    static $s;
}

// Reference assignments between other variables do not matter.
function unrelated($p, array $rows)
{
    $a = &$rows[0];
    $f = fn ($x) => $x;
    foreach ($rows as &$r) {
    }
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($p);</weak_warning>
    $p = $a;
}
