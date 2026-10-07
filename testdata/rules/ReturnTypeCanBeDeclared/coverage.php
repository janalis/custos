<?php
namespace App\Model {
    use App\Model\Parts\{Wheel, function helper, Seat as Chair};
    use App\Model\Parts\Engine;
    use function strlen;

    final class DocNormalised {
        /** @return int[] */
        public function <weak_warning descr="Declare ': array' as the return type.">ids</weak_warning>() { return []; }
        /** @return boolean */
        public function <weak_warning descr="Declare ': bool' as the return type.">flag</weak_warning>() { return $this->x(); }
        /** @return true */
        public function <weak_warning descr="Declare ': bool' as the return type.">yes</weak_warning>() { return true; }
        /** @return integer */
        public function <weak_warning descr="Declare ': int' as the return type.">count</weak_warning>() { return $this->x(); }
        /** @return \Closure */
        public function <weak_warning descr="Declare ': callable' as the return type.">cb</weak_warning>() { return function () { return 1; }; }
        /** @return $this */
        public function <weak_warning descr="Declare ': self' as the return type.">me</weak_warning>() { return ($this); }
        public function <weak_warning descr="Declare ': static' as the return type.">fresh</weak_warning>() { return new static(); }
        public function <weak_warning descr="Declare ': Wheel' as the return type.">wheel</weak_warning>() { return new Parts\Wheel(); }
        public function <weak_warning descr="Declare ': Chair' as the return type.">chair</weak_warning>() { return new Parts\Seat(); }
        public function <weak_warning descr="Declare ': Engine' as the return type.">engine</weak_warning>() { return new Parts\Engine(); }
        public function <weak_warning descr="Declare ': ?\Generator' as the return type.">gen</weak_warning>() { yield 1; }
        public function <weak_warning descr="Declare ': void' as the return type.">fails</weak_warning>() { throw new \LogicException(); }
        public function <weak_warning descr="Declare ': void' as the return type.">nested</weak_warning>() { $c = function () { return 1; }; $c(); }
        public function <weak_warning descr="Declare ': void' as the return type.">nestedFn</weak_warning>() { function inner() { return 1; } }
        /** @return void|null */
        public function <weak_warning descr="Declare ': void' as the return type.">nothing</weak_warning>() { }
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
            public function <weak_warning descr="Declare ': \ArrayObject' as the return type.">make</weak_warning>() { return new \ArrayObject(); }
        }
    }
    $anon = new class {
        public function <weak_warning descr="Declare ': \ArrayObject' as the return type.">make</weak_warning>() { return new \ArrayObject(); }
        public function pass($v) { return $v; }
        public $p;
        public function p() { return $this->p; }
    };

    interface Shape {
        /** @return int */
        public function <weak_warning descr="Declare ': int' as the return type.">area</weak_warning>();
        /** @return iterable|null */
        public function many();
        /** @return void|null */
        public function <weak_warning descr="Declare ': void' as the return type.">nothing</weak_warning>();
    }
    abstract class Base implements Shape {}
    abstract class Square extends Base implements Shape {}

    class ParentBox {
        /** @param int $n */
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">size</weak_warning>($n, $m) { return 1; }
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">typed</weak_warning>(int $n) { return $n; }
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">bare</weak_warning>($n) { return 1; }
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">short</weak_warning>($a) { return 1; }
    }
    class ChildBox extends ParentBox {
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">size</weak_warning>($n, $m) { return $n; }
        public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">typed</weak_warning>($n) { return $n; }
        public function bare($n) { return $n; }
        public function moved($n) { $n = undefined_fn(); return $n; }
        public function short($a, $b = null) { return $b; }
        public function extra($a, $b, $c) { return $c; }
    }
}
