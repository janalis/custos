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
class Collector
{
    public function __construct(&$into, $label = '') {}
    public static function fill($label, &...$targets) {}
}
function drain_first($label, array &$queue) {}
class Wizard
{
    private $steps = ['intro', 'details', 'confirm'];
    private $order = ['c', 'b', 'a'];
    private $matches = ['x', 'y', 'z'];
    private $pair = ['x', 'y', 'z'];
    private $legacyPair = ['x', 'y', 'z'];
    private $keyed = ['x', 'y', 'z'];
    private $deep = ['x', 'y', 'z'];
    private $current = ['x', 'y', 'z'];
    private $position = ['x', 'y', 'z'];
    private $each = ['x', 'y', 'z'];
    private $alias = ['x', 'y', 'z'];
    private $boxed = ['x', 'y', 'z'];
    private $source = ['x', 'y', 'z'];
    private $loopSource = ['x', 'y', 'z'];
    private $exposed = ['x', 'y', 'z'];
    private $generated = ['x', 'y', 'z'];
    private $lazy = ['x', 'y', 'z'];
    private $short = ['x', 'y', 'z'];
    private $own = ['x', 'y', 'z'];
    private $named = ['x', 'y', 'z'];
    private $tail = ['x', 'y', 'z'];
    private $spread = ['x', 'y', 'z'];
    private $spreadTail = ['x', 'y', 'z'];
    private $collected = ['x', 'y', 'z'];
    private $selfBuilt = ['x', 'y', 'z'];
    private $unknownClass = ['x', 'y', 'z'];
    private $anonymous = ['x', 'y', 'z'];
    private $dynamic = ['x', 'y', 'z'];
    private $foreign = ['x', 'y', 'z'];
    private $userFunction = ['x', 'y', 'z'];
    private $callTime = ['x', 'y', 'z'];
    private array $hooked = ['x', 'y', 'z'] { get => $this->hooked; }

    public function __construct(&$seed = null) {}

    public function next() { return array_shift($this->steps); }

    public function sorted()
    {
        sort($this->order);
        preg_match('/a/', 'a', $this->matches);
        return $this->order;
    }

    public function targets(array $rows, $helper, $callback)
    {
        [$this->pair[0], $this->pair[1]] = [$this->pair[1], $this->pair[0]];
        list($this->legacyPair, , $ignored) = $rows;
        list('a' => $this->keyed['a']) = $rows;
        [[$this->deep], $other] = $rows;
        foreach ($rows as $this->position => $this->current) {
        }
        foreach ($rows as [$this->each]) {
        }
        $alias = &$this->alias;
        $alias[] = 1;
        $box = [&$this->boxed];
        [&$first] = $this->source;
        foreach ($this->loopSource as &$entry) {
            $entry = $first;
        }
        $lazy = function &() { return $this->lazy; };
        $short = fn &() => $this->short;
        $this->helper($rows, $this->own);
        $this->helper(out: $this->named, value: $rows);
        self::many($rows, $rows, $this->tail);
        self::many(...$this->spread);
        Collector::fill(...$this->spreadTail);
        new Collector($this->collected);
        new self($this->selfBuilt);
        new Missing($this->unknownClass);
        new class($this->anonymous) {
        };
        $callback($this->dynamic);
        $helper->consume($this->foreign);
        drain_first('x', $this->userFunction);
        drain_first(&$this->callTime, $rows);
        $partial = strlen(...);
        return new Collector;
    }

    public function &exposed() { return $this->exposed; }

    public function &generated()
    {
        yield $this->generated;
        return;
    }

    private function helper($value, &$out = null) {}
    private static function many($value, &...$outs) {}
}
