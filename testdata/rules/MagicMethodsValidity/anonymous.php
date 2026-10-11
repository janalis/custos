<?php
interface Gate
{
    public function __allows();
}
trait Storage
{
    public function __set($key, $value) {}
    public function __get($key) {}
    public function __isset($key) {}
}
$a = new class implements Gate {
    public function __allows() { return true; }
    public function <error descr="The '__' prefix is reserved for magic methods.">__extra</error>() {}
    public function <error descr="'_get' is not magic; did you mean '__get'?">_get</error>($key) {}
};
$b = new class {
    use Storage;
    public function _set($key, $value) {}
};
class Snapshot
{
    public static function <error descr="__set_state must return \Snapshot; got 'array'.">__set_state</error>(array $props): array { return $props; }
}
