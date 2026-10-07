<?php
class Repo
{
    public function find($id) { return $this->rows[$id] ?? null; }
    public function each($cb) { foreach ($this->rows as $r) { yield $cb($r); } }
    public function count(): int { return 0; }
    public function warm($id): void { $this->cache[] = $id; }
    public function log($msg) { if ($msg === '') { return; } echo $msg; }
    public function wrap($v) { return function () use ($v) { return $v; }; }
}

class CachedRepo extends Repo
{
    // parent returns a value: dropping `return` changes the result
    public function find($id) { parent::find($id); }
    // parent is a generator
    public function each($cb) { parent::each($cb); }
    // parent declares a non-void return type
    public function count(): int { parent::count(); }
    // parent returns a closure
    public function wrap($v) { parent::wrap($v); }

    // still senseless: the parent returns nothing
    public function <weak_warning descr="Method 'warm' only forwards to its parent; remove it.">warm</weak_warning>($id): void { parent::warm($id); }
    public function <weak_warning descr="Method 'log' only forwards to its parent; remove it.">log</weak_warning>($msg) { parent::log($msg); }
}
