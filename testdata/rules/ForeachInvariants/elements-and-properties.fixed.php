<?php
class Packer
{
    private $boxes = 0;
    private $count = 0;

    public function setBoxes(int $n): void
    {
        $this->boxes = $n;
    }

    public function split(array $packages): array
    {
        for ($i = 0; $i < $this->boxes; $i++) {   // limit set elsewhere, written after
            $packages[$i]['weight'] = 1;
        }
        $this->boxes = count($packages);
        return $packages;
    }

    public function header(array $items): void
    {
        for ($i = 0, $this->count = count($items); $i < $this->count; $i++) {   // the fix would drop the write
            echo $items[$i];
        }
    }

    public function before(array $items): void
    {
        $skip = function () { $this->count = 0; };
        $this->count = count($items);
        foreach ($items as $iValue) {
            echo $iValue;
        }
    }

    public function only(array $items): void
    {
        for ($i = 0; $i < $this->count; $i++) {   // value from another method
            echo $items[$i];
        }
    }
}

function trimAll(array $a): void
{
    $n = count($a);
    for ($i = 0; $i < $n; $i++) {
        $a[$i] = trim($a[$i]);
        echo $a[$i];                // reads the new value
    }
}

function shiftRight(array $a): void
{
    $n = count($a);
    for ($i = 0; $i < $n; $i++) {
        $a[$i + 1] = $a[$i] * 2;    // writes the element read next
    }
}

function swapNested(array $a, int $k): void
{
    $n = count($a);
    for ($i = 0; $i < $n; $i++) {
        $a[$i][$k] = 0;
        echo $a[$i][1];             // may be the element just written
    }
}

function appendNested(array $a): void
{
    $n = count($a);
    for ($i = 0; $i < $n; $i++) {
        $a[$i][] = 1;
        echo $a[$i][0];             // may be the element just appended
    }
}

function reuse(array $a): void
{
    foreach ($a as $i => $iValue) {
        echo $iValue['name'];
        $a[$i]['seen'] = true;      // nothing reads it afterwards
    }
}

function distinctKeys(array $a): void
{
    foreach ($a as $i => $iValue) {
        $a[$i][1] = 'x';
        echo $iValue[2];
    }
}

$box = new Packer();
$rows = [1, 2];
for ($i = 0; $i < $box->size; $i++) {   // top level: no property value
    echo $rows[$i];
}

function collect(array $a, array $acc): array
{
    for ($i = 0, $acc = []; $i < count($a); $i++) {   // the fix would drop `$acc = []`
        $acc[] = $a[$i] * 2;
    }
    return $acc;
}
