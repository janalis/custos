<?php
namespace A {
    class Helper {}
    trait Loud { public function greet($n) { return "HEY $n"; } }
    trait Quiet { public function hush() { return 1; } }
    class Base
    {
        private $own = 1;
        public function limit($n = 10) { return $n; }
        public function diff($a, $b) { return $a - $b; }
        public function id($x) { return $x; }
        public function opt($a) { return $a; }
        public function arity($a, $b = 0) { return $a; }
        public function typed($a) { return $a; }
        public function nullable(int $a) { return $a; }
        public function helper(Helper $h) { return 1; }
        public function byRef($a) { return 1; }
        public function &refReturn() { return $this->shared; }
        public function __construct(protected $shared = 1) { echo 1; }
        public function channel() { return strtolower(__CLASS__); }
        public function line() { return __LINE__; }
        public function tagged($t = __CLASS__) { return $t; }
        public function next() { static $n = 0; return ++$n; }
        public function kind() { return get_class(); }
        public function dump() { return get_object_vars($this); }
        public function greet($n) { return "hello $n"; }
        public function alias() { return 1; }
        public function ghost() { return 1; }
        public function cfg() { return $this->cfg; }
        public function lazy() { return function () { return $this->cfg; }; }
        public function same($n = 10, array $o = [], ?int $x = null, \A\Helper ...$h): int { return $n; }
        public function flag($f = \PHP_INT_MAX) { return $f; }
        public function fn() { return __FUNCTION__ . get_class($this); }
        public function memo() { return function () { static $n = 0; return ++$n; }; }
        public function hush2() { return 2; }
        protected function wide($n = 10) { return $n; }
        protected function strict($x) { return $x; }
        protected function fail(): never { throw new \RuntimeException('x'); }
        protected function nothing(): void { echo 1; }
        protected function argc($a = null) { return func_num_args(); }
        protected function &ref(array &$a) { return $a; }
    }
}
namespace B {
    class Helper {}
    class Child extends \A\Base
    {
        use \A\Quiet;
        private $cfg = 5;
        public function limit($n = 50) { return $n; }
        public function diff($b, $a) { return $a - $b; }
        public function id($x): int { return $x; }
        public function opt($a = 1) { return $a; }
        public function arity($a) { return $a; }
        public function typed(int $a) { return $a; }
        public function nullable(?int $a) { return $a; }
        public function helper(Helper $h) { return 1; }
        public function byRef(&$a) { return 1; }
        public function refReturn() { return $this->shared; }
        public function __construct(protected $shared = 1) { echo 1; }
        public function channel() { return strtolower(__CLASS__); }
        public function line() { return __LINE__; }
        public function tagged($t = __CLASS__) { return $t; }
        public function next() { static $n = 0; return ++$n; }
        public function kind() { return get_class(); }
        public function dump() { return get_object_vars($this); }
        public function cfg() { return $this->cfg; }
        public function lazy() { return function () { return $this->cfg; }; }
        public function wide($n = 50) {
            return parent::wide($n);
        }
        public function strict($x): int {
            return parent::strict($x);
        }
        public function fail(): never {
            parent::fail();
        }
        public function nothing(): void {
            parent::nothing();
        }
        public function argc($a = null) { return func_num_args(); }
        public function &ref(array &$a) { return $a; }
    }
    class Traited extends \A\Base
    {
        use \A\Loud;
        use \A\Quiet { hush as alias; }
        public function greet($n) { return "hello $n"; }
        public function alias() { return 1; }
        }
    class Unknown extends \A\Base
    {
        use Missing;
        public function ghost() { return 1; }
    }
}
