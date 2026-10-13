<?php
class Stat
{
    public float $ratio = 0.0;
    public int $count = 0;
    public static int $total = 0;

    public function reset(string $raw, array $m, $o, $i, $k)
    {
        // A copied value would skip the conversion done by the inner target.
        <weak_warning descr="Split this chained assignment into separate assignments.">$shown = $this->ratio = 0</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$n = $this->count = $raw</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$n = self::$total = '7'</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$t = $raw[0] = 'xyz'</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $this->ratio = $b = 1</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $this->ratio = $b = load()</weak_warning>;
        // The outermost target is evaluated before the inner write.
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[$i++] = $b = 0</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[$o->p] = $b = 5</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[$k['i']] = $b = 7</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[key()] = $b = load()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$o->next()->p = $b = 1</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[$i++][] = $b = 1</weak_warning>;
        // A value with side effects is evaluated once.
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $b = ${name()}</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$c = $d = make()::N</weak_warning>;
        // Still split.
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $this->ratio = null</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $m[$i++] = \NULL</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $_POST['k'] = 'v'</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$this->count = $b = 3</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[$i] = $i = 9</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m[] = $b = $raw</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$m['k'][] = $a = $b = load()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">self::$total = $a = $b = 2</weak_warning>;
    }
}
