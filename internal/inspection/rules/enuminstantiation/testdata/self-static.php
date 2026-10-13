<?php
enum Stage: string {
    case Draft = 'draft';
    public static function initial(): self { return <error descr="Use an enum case instead of instantiation.">new self()</error>; }
    public static function late(): static { return <error descr="Use an enum case instead of instantiation.">new STATIC</error>; }
    public function viaClosure(): \Closure { return function () { return new self(); }; }
    public function viaArrow(): \Closure { return fn () => new static(); }
    public function viaAnonymous(): object { return new class { function again() { return new self(); } }; }
    public function other(): object { return new \stdClass(); }
}
trait Builds { public static function build(): static { return new static(); } }
class Plain { public static function build(): self { return new self(); } }
function loose() { return new self(); }
