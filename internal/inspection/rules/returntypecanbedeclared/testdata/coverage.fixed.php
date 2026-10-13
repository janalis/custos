<?php
namespace App\Model {
    use App\Model\Parts\{Wheel, function helper, Seat as Chair};
    use App\Model\Parts\Engine;
    use function strlen;

    final class DocNormalised {
        /** @return int[] */
        public function ids(): array { return []; }
        /** @return boolean */
        public function flag() { return $this->x(); }
        /** @return true */
        public function yes(): bool { return true; }
        /** @return integer */
        public function count() { return $this->x(); }
        /** @return \Closure */
        public function cb(): callable { return function () { return 1; }; }
        /** @return $this */
        public function me(): self { return ($this); }
        public function fresh(): static { return new static(); }
        public function wheel(): Wheel { return new Parts\Wheel(); }
        public function chair(): Chair { return new Parts\Seat(); }
        public function engine(): Engine { return new Parts\Engine(); }
        public function gen(): \Generator { yield 1; }
        public function fails(): void { throw new \LogicException(); }
        public function nested(): void { $c = function () { return 1; }; $c(); }
        public function nestedFn(): void { function inner() { return 1; } }
        /** @return void|null */
        public function nothing(): void { }
        /** @return int|string|null */
        public function three() { return 1; }
        public function call() { return undefined_fn(); }
        public function local() { $v = undefined_fn(); return $v; }
        public function nsafe() { return $this?->x; }
        public function other(DocNormalised $o) { return $o->x; }
    }
}
namespace App\Model\Parts {
    class Wheel {}
    class Seat {}
    class Engine {}
}
namespace {
    if (true) {
        class Conditional {
            public function make(): \ArrayObject { return new \ArrayObject(); }
        }
    }
    $anon = new class {
        public function make(): \ArrayObject { return new \ArrayObject(); }
        public function pass($v) { return $v; }
        public $p;
        public function p() { return $this->p; }
    };

    interface Shape {
        /** @return int */
        public function area(): int;
        /** @return iterable|null */
        public function many();
        /** @return void|null */
        public function nothing(): void;
    }
    abstract class Base implements Shape {}
    abstract class Square extends Base implements Shape {}

    class ParentBox {
        /** @param int $n */
        public function size($n, $m) { return 1; }
        public function typed(int $n) { return $n; }
        public function bare($n) { return 1; }
        public function short($a) { return 1; }
    }
    class ChildBox extends ParentBox {
        public function size($n, $m) { return $n; }
        public function typed($n) { return $n; }
        public function bare($n) { return $n; }
        public function moved($n) { $n = undefined_fn(); return $n; }
        public function short($a, $b = null) { return $b; }
        public function extra($a, $b, $c) { return $c; }
    }
}
