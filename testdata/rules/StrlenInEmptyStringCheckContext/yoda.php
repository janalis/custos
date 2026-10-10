<?php
function probe(string $name, ?string $alias, int $count, $raw, array $bag)
{
    if (<weak_warning descr="Compare with an empty string instead: ''' !== $name'.">strlen($name)</weak_warning>) {}
    while (!mb_strlen($name, 'UTF-8')) {}
    if (!(<weak_warning descr="Compare with an empty string instead: ''' !== (string)$alias'.">strlen($alias)</weak_warning>)) {}
    $a = mb_strlen($raw) ? 'y' : 'n';
    $b = $count > 2 and <weak_warning descr="Compare with an empty string instead: ''' !== trim($raw)'.">strlen(trim($raw))</weak_warning>;

    $c = 0 === mb_strlen($name);
    $d = <weak_warning descr="Compare with an empty string instead: ''' !== (string)$count'.">strlen($count) <> 0</weak_warning>;
    $e = <weak_warning descr="Compare with an empty string instead: ''' === (string)$alias'.">strlen($alias)  ==  0</weak_warning>;
    $f = mb_strlen($name) > 0;
    $g = <weak_warning descr="Compare with an empty string instead: ''' !== (string)$bag['k']'.">strlen($bag['k']) >= 1</weak_warning>;
    $h = <weak_warning descr="Compare with an empty string instead: ''' === $name'.">strlen($name) < 1</weak_warning>;

    $i = 0 < strlen($name);
    $j = strlen($name) > 1;
    $k = (strlen($name)) > 0;
    $l = strlen($name) ?: 5;
    return strlen($name);
}

if (<weak_warning descr="Compare with an empty string instead: ''' !== (string)$argv[1]'.">strlen($argv[1])</weak_warning>) {}
