<?php
trait Routes { private $fromTrait = []; }
class Base {
    use Routes;
    protected $routes = [];
}

class Router extends Base {
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$verbs</weak_warning> = ['GET', 'POST', 'PUT', 'DELETE'];
    protected <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$matrix</weak_warning> = array('a' => [1], 'b' => [2], 7, 'c' => "x");
    private array <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$mixed</weak_warning> = [1, 'one', 2, <<<TXT
two
TXT, 3, ['three']];
    private $first = ['a'], <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$second</weak_warning> = ["a$x", 'b', ('c')];
}
class ReadOnlyTable
{
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$names</weak_warning> = ['a', 'b', 'c'];

    public function __construct()
    {
        $copy = $this->names;
        $copy[] = 'd';
        $this->count++;
    }
}
class Sink
{
    public function __construct($byValue, &$byRef = null) {}
}
abstract class ReadOnlyUses
{
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$searched</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$passed</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$spread</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$named</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$beforeTail</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$built</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$unpacked</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$looped</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$listed</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$closed</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$nested</weak_warning> = ['a', 'b', 'c'];
    private <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$returned</weak_warning> = ['a', 'b', 'c'];

    public function run($value, array $rest)
    {
        $found = in_array($value, $this->searched, true) || count($this->searched) > 3;
        $this->take($this->passed, $rest);
        $this->lead($rest, ...$this->spread);
        $this->take(value: $this->named);
        self::tail($rest, $this->beforeTail);
        new Sink($this->built);
        [$first] = $this->unpacked;
        foreach ($this->looped as $key => $item) {
            $found = $found || $item === $first;
        }
        $copy = [$this->listed, &$rest];
        $get = function () { return $this->closed; };
        $arrow = fn () => $this->closed;
        $refs = function &() use (&$rest) {
            $inner = function () { return $this->nested; };
            $short = fn () => $this->nested;
            return $rest;
        };
        return $this->returned;
    }

    private function take($value, &$out = null) {}
    private function lead(&$first, $second = null) {}
    private static function tail(&$first, $second, &...$others) {}
    abstract protected function &source();
}
