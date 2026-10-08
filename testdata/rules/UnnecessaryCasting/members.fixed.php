<?php
class Base {
    public static function make(): int { return 1; }
    public static function loose() { return 1; }
}

class Gauge extends Base {
    private int $typed = 0;
    /** @var integer */
    private $docInt = 1;
    private $plain;

    public static function self(): string { return ''; }

    public function read(string $m, string $p, Gauge|Base $either) {
        return [
            $this->typed,
            (int) $this->docInt,
            (string) $this->plain,
            (int) $this->$p,
            (int) $either->typed,
            (int) $this->$m(),
            (int) $either->make(),
            self::self(),
            static::self(),
            parent::make(),
            Base::make(),
            (int) Base::loose(),
            (int) Base::$m(),
            (int) $p::make(),
            (int) Missing::make(),
        ];
    }
}

class Orphan {
    public function read() {
        return (int) parent::make();
    }
}

function outside() {
    return (string) self::self();
}
