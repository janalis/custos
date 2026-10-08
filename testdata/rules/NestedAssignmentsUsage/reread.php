<?php
class Holder
{
    public static $last;
    public $items = [];
    public $name;

    public function collect(array $m, $i, $k)
    {
        // The innermost target cannot be read back without side effects.
        <weak_warning descr="Split this chained assignment into separate assignments.">$target = $this->items[] = sys_get_temp_dir() . '/a.zip'</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$a = $m[$i++] = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$b = $m[g()] = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$c = $this->{$k} = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$d = $$k = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$e = $m[$$k] = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$e = $this->next()->name = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$e = $k::$last = f()</weak_warning>;
        // Re-readable targets keep the fix.
        <weak_warning descr="Split this chained assignment into separate assignments.">$x = $m['k'][$i][PHP_EOL][self::class] = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$y = $this->name = f()</weak_warning>;
        <weak_warning descr="Split this chained assignment into separate assignments.">$z = self::$last = f()</weak_warning>;
        return [$target, $a, $b, $c, $d, $e, $x, $y, $z];
    }

    private function next(): self { return $this; }
}
