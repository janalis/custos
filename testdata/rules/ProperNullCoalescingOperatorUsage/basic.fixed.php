<?php
interface Shape {}
class Square implements Shape {}
class Disc implements Shape {}
class Engine {}
class Turbo extends Engine {}

class Garage {
    /** @var Engine|null */
    private $engine;

    public function lookup(): ?Engine { return null; }

    public function demo(Engine $e, Turbo $t, Square $s, Disc $d, ?int $n = null, string $label = null) {
        return [
            $this->lookup(),
            strrev($label),
            $this->engine ?? 'none',
            $n ?? 'n/a',
            $this->engine ?? null,
            $label ?? null,
            $t ?? $e,
            $e ?? $t,
            $s ?? $d,                       // related through Shape
            $n ?? 42,
            (int) ($label ?? $e),
            $a ?? $b ?? $c,
        ];
    }
}

$top = $undefined ?? 'x';
