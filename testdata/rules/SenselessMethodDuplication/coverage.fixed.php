<?php
namespace A {
    const KNOWN = 1;
    class G { public function up() { return 1; } }
    class P extends G
    {
        public function up() { return parent::up(); }
        public function consts() { return [true, \A\KNOWN, \PHP_EOL, PHP_INT_MAX]; }
        public function calls($f) { strlen('x'); return $f(); }
        public function ghostFn() { return ghost_fn(); }
        public function ghostConst() { return GHOST_CONST; }
        public function nsConst() { return X; }
        public function mixed() { return new self() instanceof P; }
        public function refs($e) { try { return \A\P::$s . \A\P::C; } catch (\Exception $e) { return $e instanceof \Throwable; } }
        public function privStatic() { return $this::hide(); }
        public function privConst() { return $this::SECRET; }
        public function privProp() { return $this::$stash; }
        public function privLater() { $this->hide(); return 1; }
        public function dyn($n) { return $this->$n(); }
        protected function typed(): int { return 1; }
        protected function docRet() { return 1; }
        protected function gen() { yield 1; }
        protected function closure() { $f = function () { return 1; }; echo $f(); }
        protected function early($a) { if ($a) { return 1; } echo 2; }
        public static $s = '';
        const C = 1;
        private const SECRET = 1;
        private static $stash = 1;
        private static function hide() { return 1; }
        public function count() { return 1; }
    }
}
namespace B {
    const X = 1;
    class C extends \A\P
    {
        public function up() { return parent::up(); }
        public function ghostFn() { return ghost_fn(); }
        public function ghostConst() { return GHOST_CONST; }
        public function nsConst() { return X; }
        public function mixed() { return new self() instanceof \A\P; }
        public function privStatic() { return $this::hide(); }
        public function privConst() { return $this::SECRET; }
        public function privProp() { return $this::$stash; }
        public function privLater() { $this->hide(); return 1; }
        public function typed(): int {
            return parent::typed();
        }
        /** @return int */
        public function docRet() {
            return parent::docRet();
        }
        public function gen() {
            return parent::gen();
        }
        public function closure() {
            parent::closure();
        }
        public function early($a) {
            return parent::early($a);
        }
    }
    class Arr extends \ArrayObject
    {
        public function count(): int { return 1; }
    }
    class FooTest extends \A\P
    {
        public function up() { return parent::up(); }
    }
    $anon = new class extends \A\P {
        public function up() { return parent::up(); }
    };
}
