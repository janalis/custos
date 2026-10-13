<?php
class Account {
    private $owner;
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

class Hooks {
    /* @var callable */
    protected $onSave;
    /** @var \Closure|null */
    protected $onLoad;
    protected \Closure $onRun;
    public function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onSave</weak_warning>() {}
    public static function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onLoad</weak_warning>() {}
    private function <weak_warning descr="A callable property with this name exists; rename the method (for example with a get/is/has prefix).">onRun</weak_warning>() {}
}

class Child extends Account {
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">owner</weak_warning>() {}
}

trait Bag {
    public static $items;
}

class UsesBag {
    use Bag;
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">items</weak_warning>() {}
}
