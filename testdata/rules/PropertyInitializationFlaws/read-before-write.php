<?php
class Registry
{
    private $items = ['core'];
    private $count = 10;
    private $label = 'base';
    private $seen = 'never';
    private $buf = [1];
    private $mode = <weak_warning descr="Default is always replaced by the constructor; remove it.">'auto'</weak_warning>;
    private $limit = <weak_warning descr="Default is always replaced by the constructor; remove it.">5</weak_warning>;

    public function __construct(array $extra, $strict)
    {
        if ($strict) {
            $this->mode = 'strict'; // a nested plain write reads nothing
        }
        $this->mode = 'manual';
        $this->limit = count($extra);
        $merged = $this->items; // direct read of the default
        foreach ($extra as $e) {
            $merged[] = $e;
        }
        $this->items = $merged;
        $this->count += 5; // compound write uses the default
        $this->count = count($extra);
        echo "label was {$this->label}\n";
        $this->label = 'custom';
        $f = function () { return $this->seen; };
        $was = $f();
        $this->seen = 'now:' . $was;
        $this->buf[] = 2; // modifies the default in place
        $this->buf = [count($extra)];
    }
}

class Snapshot
{
    private $state = ['new'];
    private $flags = [];
    private $name = <weak_warning descr="Default is always replaced by the constructor; remove it.">'none'</weak_warning>;

    public function __construct(array $state, array $flags, $key)
    {
        $this->name = 'snap';
        $copy = (array) $this; // sees every default
        $this->state = $state + $copy;
        $this->flags = $flags;
    }
}

class Lookup
{
    private $table = ['a' => 1];
    private $extra = [];

    public function __construct(array $table, $key)
    {
        $first = $this->{$key}; // dynamic name: any property
        $this->table = $table;
        $this->extra = [$first];
    }
}
