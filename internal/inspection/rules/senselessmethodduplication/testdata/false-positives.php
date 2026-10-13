<?php
namespace A {
    class Helper {}
    abstract class Base
    {
        public function make() { return new Helper(); }
        public function self() { return new self(); }
        public function ghost() { return new Ghost(); }
        abstract public function todo();
        /** @deprecated */
        public function old() { return 1; }
        public function fresh() { return 1; }
        public function empty() {}
        private function mine() { return 1; }
        public function differs() { echo 1; }
        public function more() { echo 1; }
    }
}
namespace B {
    class Helper {}
    class Child extends \A\Base
    {
        public function make() { return new Helper(); }
        public function self() { return new self(); }
        public function ghost() { return new Ghost(); }
        public function todo() { return 1; }
        public function old() { return 1; }
        #[\Deprecated]
        public function fresh() { return 1; }
        public function empty() {}
        private function mine() { return 1; }
        public function differs() { echo 2; }
        public function more() { echo 1; echo 1; }
        public function own() { echo 1; }
    }
    trait T { public function more() { echo 1; } }
    interface I { public function more(); }
}
