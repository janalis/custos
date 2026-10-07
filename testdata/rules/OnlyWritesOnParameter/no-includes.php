<?php
class Report
{
    public function collect(array $rows, array &$sink, $tally, Closure $done, Iterator $it)
    {
        return function ($pending, array &$bucket) use (<weak_warning descr="Variable is never used.">$rows</weak_warning>, &<weak_warning descr="Variable is never used.">$sink</weak_warning>, $tally) {
            <weak_warning descr="Value is only written here and never read; the write is lost.">$pending['x']</weak_warning> = 1;
            $bucket[] = 2;
            <weak_warning descr="Value is only written here and never read; the write is lost.">$tally[]</weak_warning> = 3;
        };
    }

    public function counters($hits, $misses, $sum, &$total, Iterator $it)
    {
        <weak_warning descr="Value is only written here and never read; the write is lost.">$hits++</weak_warning>;
        <weak_warning descr="Value is only written here and never read; the write is lost.">--$misses</weak_warning>;
        <weak_warning descr="Value is only written here and never read; the write is lost.">$sum</weak_warning> *= 2;
        $total++;
        $it[] = 1;
    }

    public function inline_unused()
    {
        if (null !== (<weak_warning descr="Variable is never used.">$pos</weak_warning> = strpos('abc', 'b'))) {
            return true;
        }
        $map = [];
        $map[<weak_warning descr="Variable is never used.">$key</weak_warning> = 'k'] = 1;
        $map[$slot = 'z'] = 2;
        return $slot . count($map);
    }

    public function locals($flag)
    {
        $buffer = $other = [];
        if ($flag) {
            <weak_warning descr="Value is only written here and never read; the write is lost.">$buffer[]</weak_warning> = 1;
        }
        <weak_warning descr="Value is only written here and never read; the write is lost.">$other</weak_warning> .= 'x';

        $kept = '';
        $kept .= 'y';
        echo $kept ?? '';

        $alias = &$flag;
        $alias .= 'z';

        $unusedPlain = 5;
    }

    public function with_include($n, $path)
    {
        $n -= 1;
        require $path;
    }
}

function silenced(array $out)
{
    /** Buffers. @noinspection OnlyWritesOnParameterInspection */
    $out['a'][] = 1;
    $out['b'][] = 2;
}
