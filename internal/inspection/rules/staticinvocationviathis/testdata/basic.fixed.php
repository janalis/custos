<?php
class Palette {
    public static function shade($level) { return $level; }
    public static function staticTone() { return 0; }
    public function tint() { return 1; }

    public static function fromStatic() {
        return self::shade(1) + $this->shade(1);
    }

    public function render() {
        $a = static::shade(2);
        $a2 = static :: SHADE(2);
        $b = $this->tint();
        $c = $this->staticTone();
        $d = self::shade(3);
        $e = function () { return $this->shade(4); };
        $f = fn() => $this->shade(4);
        $g = $this?->shade(4);
    }
}

function paint(Palette $given) {
    $local = new Palette();
    $x = $local->shade(5);
    $y = (new Palette())->shade(6);
    $z = $given->shade(7);
    $w = makePalette()->shade(8);
    return function () use ($local) { return $local->shade(9); };
}

$top = new Palette();
$top->shade(10);
