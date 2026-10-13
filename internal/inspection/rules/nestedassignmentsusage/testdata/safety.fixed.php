<?php
class Stat
{
    public float $ratio = 0.0;
    public int $count = 0;
    public static int $total = 0;

    public function reset(string $raw, array $m, $o, $i, $k)
    {
        // A copied value would skip the conversion done by the inner target.
        $shown = $this->ratio = 0;
        $n = $this->count = $raw;
        $n = self::$total = '7';
        $t = $raw[0] = 'xyz';
        $a = $this->ratio = $b = 1;
        $a = $this->ratio = $b = load();
        // The outermost target is evaluated before the inner write.
        $m[$i++] = $b = 0;
        $m[$o->p] = $b = 5;
        $m[$k['i']] = $b = 7;
        $m[key()] = $b = load();
        $o->next()->p = $b = 1;
        $m[$i++][] = $b = 1;
        // A value with side effects is evaluated once.
        $b = ${name()};
        $a = $b;
        $d = make()::N;
        $c = $d;
        // Still split.
        $this->ratio = null;
        $a = null;
        $m[$i++] = \NULL;
        $a = \NULL;
        $_POST['k'] = 'v';
        $a = 'v';
        $b = 3;
        $this->count = 3;
        $i = 9;
        $m[$i] = 9;
        $b = $raw;
        $m[] = $raw;
        $b = load();
        $a = $b;
        $m['k'][] = $b;
        $b = 2;
        $a = 2;
        self::$total = 2;
    }
}
