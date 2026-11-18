<?php

class Zone {}

/** @param resource|null $handle */
function scalars($handle, ?int $count, ?bool $on)
{
    return [
        empty($handle),
        empty($count),
        !empty($on),
    ];
}

function zone(mixed $tz, bool $flag)
{
    if ($tz instanceof Zone) {
        $maybe = $tz;
    }
    $sure = new Zone();
    global $shared;
    static $kept = null;
    if ($flag) {
        return empty($maybe);
    }
    return [
        $sure === null,
        $shared, $kept,
    ];
}

class Holder
{
    public function probe()
    {
        global $other;
        /** @var Zone $g */
        global $g;
        /** @var Zone $s */
        static $s;
        return [
            $this === null,
            empty($g),
            empty($s),
            $other,
        ];
    }
}

function imports(?Zone $p)
{
    $later = static function () use ($p) {
        $fresh = new Zone();
        return [
            $p === null,
            $fresh === null,
        ];
    };
    return $p === null || $later;
}

function counted(bool $flag)
{
    if ($flag) {
        $rows = [1];
    }
    return empty($rows);
}
