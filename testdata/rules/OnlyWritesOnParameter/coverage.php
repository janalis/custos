<?php
function indexRead($p)
{
    $x = $p['k'];
    return $x;
}

function indexIncrement($p)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$p['k']</weak_warning>++;
}

function indexEcho($p)
{
    echo $p['k'];
}

function incAfterRef($o)
{
    $v = &$o;
    $v++;
    $v = 2;
    return $o;
}

function incConsumed($p)
{
    return $p++;
}

function unaryRead($p, $q)
{
    -$q;
    return !$p;
}

function compound($p, $q)
{
    $x = 'a';
    $x .= $q;
    return [$p .= 'x', $x];
}

function comparisons()
{
    if (false === $found = lookup()) {
        return null;
    }
    if (true && $other = lookup()) {
        return $other;
    }
    return $found;
}

function caught()
{
    $e = null;
    try {
        work();
    } catch (\Exception <weak_warning descr="Value is only written here and never read; the write is lost.">$e</weak_warning>) {
    }
}

function keys($p, $q)
{
    $a = [$p => 1, [$q] => 2];
    return $a;
}

function consumedIndexWrite($p)
{
    foo(<weak_warning descr="Value is only written here and never read; the write is lost.">$p['k']</weak_warning> = 1);
}

function objectImport(object $o)
{
    return function () use ($o) {
        $o['k'] = 1;
    };
}
