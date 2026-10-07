<?php
class Store
{
    public function fill(&$target) { $target = []; }
    public static function make() { return []; }

    public function shapes($m, $cls)
    {
        $this->fill($this->$m());
        $this->fill(Store::$m());
        $this->fill($cls::make());
        $this->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">self::make()</warning>);
        $this->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">static::make()</warning>);
        $this->fill(parent::make());
    }
}

class Shelf extends Store
{
    public function more()
    {
        $this->fill(<warning descr="Pass a variable here: this parameter is taken by reference.">parent::make()</warning>);
    }
}

function grab(&$x) { $x = 1; }
$anon = new class {
    public function f() { grab(parent::make()); }
};
