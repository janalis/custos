<?php
$a = new class {
    const LIMIT = 1;
    public $handler;
    public int $count = 0;

    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">handler</weak_warning>() {}
    public function count() {}
    public function other() {}
};

// Inherited from the PHP stubs (no declaration in this file): typed by its
// stub declaration and default.
class AppException extends \Exception
{
    public function message() {}
    public function trace() {}
}

// Error recovery: a method without a name.
class Broken
{
    public $x;
    public function () {}
}
