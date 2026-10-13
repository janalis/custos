<?php
namespace A { class Widget {} class Base { protected $kind = Widget::class; } }
namespace B {
    class Widget {}
    class Child extends \A\Base { protected $kind = Widget::class; }
    class Typed
    {
        private ?Widget $w = null;
        private mixed $any = null;
        private int|null $u = null;
        private ?int $n;
        public function __construct($seed = 0, private $promoted = null) { $this->n = null; }
    }
    trait T { private $x = 1; public function __construct() { $this->x = 1; } }
}
