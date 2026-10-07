<?php
function collect(array $groups, $limit) {
    for ($n = 0; $n < $limit; ++$n) {
        do {
            <warning descr="Array '$bucket' is never initialised; initialise it before the loops.">$bucket[]</warning> = $n;
            <warning descr="Array '$bucket' is never initialised; initialise it before the loops.">$bucket[$n][]</warning> = $limit;
        } while (false);
    }

    foreach ($groups as $group) {
        $flat[] = $group;
    }

    $done = [];
    foreach ($groups as $group) {
        foreach ($group as $item) {
            $done[] = $item;
            $limit[] = $item;
            $group[] = 0;
            $this->seen[] = $item;
        }
    }

    return function ($extra) use ($done) {
        while (true) {
            foreach ($extra as $e) {
                $done[] = $e;
                $extra[] = $e;
                <warning descr="Array '$fresh' is never initialised; initialise it before the loops.">$fresh[]</warning> = count($fresh);
            }
        }
    };
}

class Repo
{
    public function rows($data)
    {
        foreach ($data as $row) {
            while ($row) {
                <warning descr="Array '$out' is never initialised; initialise it before the loops.">$out['k'][]</warning> = $row;
            }
        }
        return $out;
    }
}

foreach ($a as $x) {
    foreach ($x as $y) {
        $global[] = $y;
    }
}
