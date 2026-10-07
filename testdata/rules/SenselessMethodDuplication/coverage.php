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
        public function <weak_warning descr="Method 'consts' duplicates the inherited implementation; remove it.">consts</weak_warning>() { return [true, \A\KNOWN, \PHP_EOL, PHP_INT_MAX]; }
        public function <weak_warning descr="Method 'calls' duplicates the inherited implementation; remove it.">calls</weak_warning>($f) { strlen('x'); return $f(); }
        public function ghostFn() { return ghost_fn(); }
        public function ghostConst() { return GHOST_CONST; }
        public function nsConst() { return X; }
        public function mixed() { return new self() instanceof \A\P; }
        public function <weak_warning descr="Method 'refs' duplicates the inherited implementation; remove it.">refs</weak_warning>($e) { try { return \A\P::$s . \A\P::C; } catch (\Exception $e) { return $e instanceof \Throwable; } }
        public function privStatic() { return $this::hide(); }
        public function privConst() { return $this::SECRET; }
        public function privProp() { return $this::$stash; }
        public function privLater() { $this->hide(); return 1; }
        public function <weak_warning descr="Method 'dyn' duplicates the inherited implementation; remove it.">dyn</weak_warning>($n) { return $this->$n(); }
        public function <weak_warning descr="Method 'typed' duplicates the inherited implementation; delegate to parent::typed() instead.">typed</weak_warning>(): int { return 1; }
        /** @return int */
        public function <weak_warning descr="Method 'docRet' duplicates the inherited implementation; delegate to parent::docRet() instead.">docRet</weak_warning>() { return 1; }
        public function <weak_warning descr="Method 'gen' duplicates the inherited implementation; delegate to parent::gen() instead.">gen</weak_warning>() { yield 1; }
        public function <weak_warning descr="Method 'closure' duplicates the inherited implementation; delegate to parent::closure() instead.">closure</weak_warning>() { $f = function () { return 1; }; echo $f(); }
        public function <weak_warning descr="Method 'early' duplicates the inherited implementation; delegate to parent::early() instead.">early</weak_warning>($a) { if ($a) { return 1; } echo 2; }
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
