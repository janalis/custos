<?php
trait Routes { private $fromTrait = []; }
class Base {
    use Routes;
    protected $routes = [];
}

class Router extends Base {
    protected $routes = ['a', 'b', 'c'];
    private $fromTrait = ['a', 'b', 'c'];
    private $pair = ['x', 'y'];
    private $nested = [['a', 'b', 'c', 'd']];
    public $open = ['p', 'q', 'r'];
    var $legacy = ['p', 'q', 'r'];
    private static $shared = ['p', 'q', 'r'];
    private $numbers = [1, 2, 3, 4];
    private $spread = [...['a', 'b'], 'c', 'd'];
    private $concat = ['a' . 'b', 'c' . 'd', 'e'];
    const NAMES = ['p', 'q', 'r'];
    public function __construct(private $promoted = ['a', 'b', 'c']) {}
}
