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
