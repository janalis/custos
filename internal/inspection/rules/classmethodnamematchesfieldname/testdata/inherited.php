<?php
interface Transformer { public function transform($v); }
abstract class Listener { abstract public function onInput(); }
trait Secret { private $token; }
class Vault { use Secret; }

class Callback implements Transformer {
    private \Closure $transform;
    public function <weak_warning descr="A callable property with this name exists; the method name is inherited, so rename the property.">transform</weak_warning>($v) { return ($this->transform)($v); }
}

class Field extends Listener {
    protected $onInput;
    public function <weak_warning descr="A property with this name exists and its type is unknown; the method name is inherited, so rename or type the property.">onInput</weak_warning>() {}
}

class Counter implements \Countable {
    private $count;
    public function <weak_warning descr="A property with this name exists and its type is unknown; the method name is inherited, so rename or type the property.">count</weak_warning>(): int { return 0; }
}

class Door extends Vault {
    public function token() {}           // custos: the trait property is private to Vault
}

class Lamp extends Missing {
    private $glow;
    public function <weak_warning descr="A property with this name exists and its type is unknown; rename the method or type the property.">glow</weak_warning>() {}
}
