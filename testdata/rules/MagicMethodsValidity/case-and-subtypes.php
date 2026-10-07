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
    public static function <error descr="__ToString must not be static.">__ToString</error>() { return 'x'; }
    public static function __Set_State($p) { <error descr="__Set_State must return \Broken; got '\Shape'.">return new Shape();</error> }
    public function <error descr="The '__' prefix is reserved for magic methods.">__Fetch</error>() {}
    public function <error descr="'_Invoke' is not magic; did you mean '__Invoke'?">_Invoke</error>() {}
}

class Parent1 { public function __construct() {} }
class Child1 extends Parent1 { public function __construct() { parent::__Construct(); } }
