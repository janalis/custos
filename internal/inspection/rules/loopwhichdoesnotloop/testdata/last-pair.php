<?php
class Cursor
{
    public $key;
    public $value;

    /** @param array<string, int> $pairs */
    public function seekEnd(array $pairs, array $rows)
    {
        // Each target outlives the loop: the last pair stays.
        foreach ($pairs as $this->key => $this->value) {
        }
        $last = [];
        foreach ($pairs as $last['key'] => $last['value']) {
        }
        foreach ($rows as [$a, $b]) {
        }
        foreach ($pairs as $name => $count) {
        }
        return [$last, $name, $a, $b];
    }

    /** @param array<string, int> $pairs */
    public function unused(array $pairs)
    {
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($pairs as $name => $count) {
        }
        $f = function () { return $count; };
        return $f;
    }
}
