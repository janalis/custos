<?php
class Shape
{
    public static function __set_state($props) { return new Circle(); }
}

class Circle extends Shape
{
    public static function __Set_State($props): Circle { return new Circle(); }
}

class Label
{
    public function __ToString() { return 'x'; }
    public function __TOARRAY() { return []; }
    public function __Get($k) { return null; }
    public function __Set($k, $v) {}
    public function __ISSET($k) { return false; }
}

class Broken
{
    public static function __ToString() { return 'x'; }
    public static function __Set_State($p) { return new Shape(); }
    public function __Fetch() {}
    public function __Invoke() {}
}

class Parent1 { public function __construct() {} }
class Child1 extends Parent1 { public function __construct() { parent::__Construct(); } }
