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
            <weak_warning descr="'$this->lookup()' alone is equivalent; drop the '?? null' fallback.">$this->lookup() ?? null</weak_warning>,
            <weak_warning descr="'strrev($label)' alone is equivalent; drop the '?? null' fallback.">strrev($label) ?? NULL</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([\Engine] vs [string]).">$this->engine ?? 'none'</weak_warning>,
            <weak_warning descr="Operand types of '??' do not match ([int] vs [string]).">$n ?? 'n/a'</weak_warning>,
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
