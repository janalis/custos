<?php
// The parent constructor called through a callable (Swiftmailer).
class Base
{
    public function __construct($a) { echo $a; }
}

class ByName extends Base
{
    public function __construct() { call_user_func_array('Base::__construct', [1]); }
}

class ByParent extends Base
{
    public function __construct() { call_user_func([$this, 'parent::__construct'], 1); }
}

class ByClass extends Base
{
    public function __construct() { call_user_func_array(parent::class . '::__CONSTRUCT', [1]); }
}

class Other extends Base
{
    public function <error descr="__construct does not call Base::__construct().">__construct</error>() { call_user_func_array('Base::init', [1]); strlen('x'); call_user_func(); }
}
