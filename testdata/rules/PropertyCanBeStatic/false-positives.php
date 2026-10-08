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
class Connection
{
    protected $options = ['prefix' => '', 'start' => '"', 'end' => '"'];
    private $hooks = ['a', 'b', 'c'];
    private $counters = ['x', 'y', 'z'];
    private $flags = ['p', 'q', 'r'];
    private $state = ['on', 'off', 'idle'];

    public function __construct(string $prefix)
    {
        $this->options['prefix'] = $prefix;
        $this->hooks[] = 'd';
        ($this->counters)['x'] .= '!';
        unset($this->flags[0]);
        $this->state = [];
        $other->list = [];
        $this->{$prefix} = 1;
    }
}
