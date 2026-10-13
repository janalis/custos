<?php
// A brace starts an interpolation only when `$` follows it: `{static::m()}`
// would be plain text, so these are reported without a fix.
class Swatch {
    public static function shade($level) { return new Tone("s$level"); }
    private static function index() { return 0; }

    public function render(array $rows) {
        $a = "x {$this->shade(2)} y";
        $b = <<<TXT
        v {$this->index()} w
        TXT;
        $c = "{$this->shade(3)->name}";
        // Nested deeper, the call is an ordinary expression: fixed.
        $d = "{$rows[self::index()]}";
        return [$a, $b, $c, $d];
    }
}

class Tone {
    public $name;
    public function __construct($name) { $this->name = $name; }
    public function __toString() { return $this->name; }
}
