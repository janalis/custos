<?php
class Holder
{
    public static $last;
    public $items = [];
    public $name;

    public function collect(array $m, $i, $k)
    {
        // The innermost target cannot be read back without side effects.
        $target = $this->items[] = sys_get_temp_dir() . '/a.zip';
        $a = $m[$i++] = f();
        $b = $m[g()] = f();
        $c = $this->{$k} = f();
        $d = $$k = f();
        $e = $m[$$k] = f();
        $e = $this->next()->name = f();
        $e = $k::$last = f();
        // Re-readable targets keep the fix.
        $m['k'][$i][PHP_EOL][self::class] = f();
        $x = $m['k'][$i][PHP_EOL][self::class];
        $this->name = f();
        $y = $this->name;
        self::$last = f();
        $z = self::$last;
        return [$target, $a, $b, $c, $d, $e, $x, $y, $z];
    }

    private function next(): self { return $this; }
}
