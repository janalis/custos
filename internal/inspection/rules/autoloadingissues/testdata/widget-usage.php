<?php
class <warning descr="File name does not match the class name; autoloading may fail.">Widget</warning>
{
}

function make(): Widget
{
    return new Widget();
}

$factory = function () {
    return new class extends Widget {
        public function __construct() { new Widget(); }
    };
};
$name = Widget::class;
$other = new Gadget();
Gadget::make();
