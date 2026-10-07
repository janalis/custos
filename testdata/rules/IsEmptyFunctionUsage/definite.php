<?php

class Zone {}

/** @param resource|null $handle */
function scalars($handle, ?int $count, ?bool $on)
{
    return [
        <weak_warning descr="Replace with '$handle === null'.">empty($handle)</weak_warning>,
        <weak_warning descr="Prefer a type-specific check over empty().">empty($count)</weak_warning>,
        !<weak_warning descr="Prefer a type-specific check over empty().">empty($on)</weak_warning>,
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
        return <weak_warning descr="Prefer a type-specific check over empty().">empty($maybe)</weak_warning>;
    }
    return [
        <weak_warning descr="Replace with '$sure === null'.">empty($sure)</weak_warning>,
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
            <weak_warning descr="Replace with '$this === null'.">empty($this)</weak_warning>,
            <weak_warning descr="Replace with '$g === null'.">empty($g)</weak_warning>,
            <weak_warning descr="Replace with '$s === null'.">empty($s)</weak_warning>,
            $other,
        ];
    }
}

function imports(?Zone $p)
{
    $later = static function () use ($p) {
        $fresh = new Zone();
        return [
            <weak_warning descr="Replace with '$p === null'.">empty($p)</weak_warning>,
            <weak_warning descr="Replace with '$fresh === null'.">empty($fresh)</weak_warning>,
        ];
    };
    return <weak_warning descr="Replace with '$p === null'.">empty($p)</weak_warning> || $later;
}

function counted(bool $flag)
{
    if ($flag) {
        $rows = [1];
    }
    return <weak_warning descr="Prefer a type-specific check over empty().">empty($rows)</weak_warning>;
}
