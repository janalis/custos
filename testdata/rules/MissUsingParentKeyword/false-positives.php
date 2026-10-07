<?php
class Base { public function a() {} public function b() {} }
trait Helper { public function b() {} }
class Child extends Base {
    use Helper;
    public function a() { parent::b(); }
    public function c() { parent::a(); }
    public function A2() { parent::c(); }
}
trait T { public function x() { parent::a(); } }
