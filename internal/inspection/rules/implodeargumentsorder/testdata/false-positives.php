<?php

$ok1   = implode(' | ', $labels);
$ok2   = implode($labels, $sep);
$ok3   = implode($labels);
$ok4   = join($labels, '-');
$ok5   = App\implode($labels, '-');
$ok6   = $s->implode($labels, '-');
$ok7   = Str::implode($labels, '-');
$ok8   = implode($labels, SEP);
$ok9   = implode($a, '-', $b);
$ok10  = implode(',', 'abc');
$ok11  = implode("-", <<<TXT
a
TXT);

$join = 'implode';
echo $join($parts, ',');

$ok12  = implode(array: $labels, separator: ', ');
$ok13  = implode(separator: ', ', array: $labels);
$ok14  = implode($labels, separator: ', ');
$ok15  = implode(...$args, ...['-']);
