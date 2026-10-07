<?php
function bump(array $m, int $b, $repo)
{
    $m[$b++] = $m[$b++] + 1;        // the target is evaluated twice
    $m[$b = 2] = $m[$b = 2] . 'x';
    $f = [<weak_warning descr="Use the compound form '$m[fn () => $b++] .= &quot;y&quot;'.">$m[fn () => $b++] = $m[fn () => $b++] . "y"</weak_warning>]; // closures are not run here
    $m[$b] = <weak_warning descr="Use the compound form '$m[$b] += 1'.">$m[$b] = $m[$b] + 1</weak_warning>;
    return $m;
}
