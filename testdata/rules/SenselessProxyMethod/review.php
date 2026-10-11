<?php
namespace Review;

const DEPTH = 3;

trait Loud { public function greet($n) { return "HEY $n"; } }
trait Quiet { public function whisper($n) { return $n; } }
trait Nested { use Loud; }
trait Renamed { public function original($n) { return $n; } }
trait Wrapped { use Quiet; }

class Base
{
    const LIMIT = 10;
    public function __construct(protected $id, public readonly int $rank = 0) {}
    public function greet($n) { return "hello $n"; }
    public function page($size = self::LIMIT) { return $size; }
    public function depth($d = DEPTH, $on = true, $none = null) { return $d; }
    public function limit($l = Base::LIMIT) { return $l; }
    public function add(array $items) { return count($items); }
    public function all(...$items) { return count($items); }
    public function fill(&$out) { $out[] = 1; }
    public function &ref($a) { return $this->store[$a]; }
    public function tagged($n = [__LINE__]) { return $n; }
    public function passed($a, $b = null) { return func_num_args(); }
    public function given($a) { return func_get_args(); }
    public $store = [];
}

// A trait without a method of that name does not interfere.
class Plain extends Base
{
    use Quiet;

    public function <weak_warning descr="Method 'greet' only forwards to its parent; remove it.">greet</weak_warning>($n)
    {
        return parent::greet($n);
    }

    public function <weak_warning descr="Method 'depth' only forwards to its parent; remove it.">depth</weak_warning>($d = DEPTH, $on = true, $none = null) { return parent::depth($d, $on, $none); }
    public function <weak_warning descr="Method 'limit' only forwards to its parent; remove it.">limit</weak_warning>($l = Base::LIMIT) { return parent::limit($l); }
    public function <weak_warning descr="Method 'fill' only forwards to its parent; remove it.">fill</weak_warning>(&$out) { parent::fill($out); }
    // The proxy always passes the optional argument the caller left out.
    public function passed($a, $b = null) { return parent::passed($a, $b); }
    public function <weak_warning descr="Method 'given' only forwards to its parent; remove it.">given</weak_warning>($a) { return parent::given($a); }
}

// A trait reached twice is walked once.
class Twice extends Base
{
    use Quiet, Wrapped;

    public function <weak_warning descr="Method 'greet' only forwards to its parent; remove it.">greet</weak_warning>($n) { return parent::greet($n); }
}

// The override selects the parent's method over the trait's.
class Direct extends Base
{
    use Loud;

    public function greet($n) { return parent::greet($n); }
}

class Indirect extends Base
{
    use Nested;

    public function greet($n) { return parent::greet($n); }
}

class Aliased extends Base
{
    use Renamed { original as Greet; }

    public function greet($n) { return parent::greet($n); }
}

class Unknown extends Base
{
    use \Vendor\Missing;

    public function greet($n) { return parent::greet($n); }
}

class Child extends Base
{
    const LIMIT = 50;

    // A second promotion with the same visibility redeclares nothing new.
    public function <weak_warning descr="Method '__construct' only forwards to its parent; remove it.">__construct</weak_warning>(protected $id, public readonly int $rank = 0) { parent::__construct($id, $rank); }
    // `self` names another class here.
    public function page($size = self::LIMIT) { return parent::page($size); }
    // The variadic packs its arguments into one array.
    public function add(...$items) { return parent::add($items); }
    public function all(...$items) { return parent::all($items); }
    public function &ref($a) { return parent::ref($a); }
    public function tagged($n = [__LINE__]) { return parent::tagged($n); }
}

class PrivateBase { public function __construct(private $id) {} }
class PrivateChild extends PrivateBase
{
    // The child's own private property would disappear.
    public function __construct(private $id) { parent::__construct($id); }
}

class GuardedBase { public function __construct(public protected(set) int $id) {} }
class GuardedChild extends GuardedBase
{
    // The redeclaration widens who may write the property.
    public function __construct(public int $id) { parent::__construct($id); }
}

class ByRefBase { public function __construct(&$log) { $log[] = 'base'; } }
class ByRefChild extends ByRefBase
{
    public function __construct($log) { parent::__construct($log); }
}
