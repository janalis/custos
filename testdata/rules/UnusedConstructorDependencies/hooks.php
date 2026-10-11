<?php
// A property hook runs after construction: what it reads is not dead.
class Person
{
    private string $first;
    private string $last;
    private string $unused;

    public string $fullName {
        get => $this->first . ' ' . $this->last;
    }

    public function __construct(string $first, string $last)
    {
        $this->first = $first;
        $this->last = $last;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->unused</weak_warning> = $first;
    }
}

// Hooks of promoted constructor parameters count as well.
class Promoted
{
    private $prefix;
    private $suffix;

    public function __construct(
        $prefix,
        public string $label { get => $this->prefix . $this->label; },
    ) {
        $this->prefix = $prefix;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->suffix</weak_warning> = $prefix;
    }
}

// A hook reading properties by a computed name may read any of them.
class Lookup
{
    private $a;

    public array $keys = [] {
        set(array $value) { $this->keys = array_map(fn($k) => $this->$k, $value); }
    }

    public function __construct($a) { $this->a = $a; }
}

trait Titled
{
    public string $title { get => $this->name; }
}

class Page
{
    use Titled;

    private $name;

    public function __construct($name) { $this->name = $name; }
}

trait Exports
{
    public array $all { get => get_object_vars($this); }
}

class Record
{
    use Exports;

    private $id;

    public function __construct($id) { $this->id = $id; }
}

abstract class Shape
{
    private $sides;

    abstract public int $area { get; }

    public function __construct($sides) { <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->sides</weak_warning> = $sides; }
}

// Assigning a hooked property runs its set hook: not a plain dead store.
class Audited
{
    public array $log = [];
    private string $name { set { $this->log[] = $value; } }
    private $plain;

    public function __construct($name)
    {
        $this->name = $name;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->plain</weak_warning> = $name;
    }
}
