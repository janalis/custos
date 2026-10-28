<?php
// Reads go through __get and __isset; there is nothing to write.
final class View
{
    public function __construct(private array $data) {}
    public function __get($key) { return $this->data[$key]; }
    public function __isset($key) { return isset($this->data[$key]); }
}

abstract class BaseView
{
    public function <error descr="__get needs a companion __set method.">__get</error>($key) { return null; }
}

final class ChildView extends BaseView
{
    public function __isset($key) { return false; }
}

final class Removable
{
    public function <error descr="__get needs a companion __set method.">__get</error>($key) { return null; }
    public function <error descr="__isset needs a companion __set method.">__isset</error>($key) { return false; }
    public function <error descr="__unset needs a companion __set method.">__unset</error>($key) {}
}

final class GetOnly
{
    public function <error descr="__get needs a companion __set method.">__get</error>($key) { return null; }
}
