<?php
namespace Review;

interface Filter
{
    public function __isAllowed();
}
abstract class Cache
{
    abstract protected function _get($key);
}
trait Hooks
{
    abstract public function _set($key, $value);
}

// Anonymous classes: the name is imposed by the interface, parent or trait.
$filter = new class implements Filter {
    public function __isAllowed() { return true; }
};
$cache = new class extends Cache {
    protected function _get($key) { return 1; }
};
$hooks = new class {
    use Hooks;
    public function _set($key, $value) {}
};
// Anonymous class declaring the magic method next to its hook.
$model = new class {
    public function __construct() {}
    protected function _construct() {}
};

// `object` is the documented return type of __set_state.
class State
{
    public static function __set_state(array $props): object { return new self(); }
}
class Loose
{
    public static function __set_state(array $props) { return self::build($props); }
    private static function build(array $props): object { return new self(); }
}
