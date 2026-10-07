<?php
class Store
{
    protected function __construct($cfg = []) {}

    public function put($key, $value = 0) { return true; }

    public function get($key) {}

    public function drop($key) {}

    public function tag($key) {}

    /**
     * @api
     */
    public function flush($all) {}

    public function label($text = __CLASS__) {}

    public function size(): int { return 0; }

    public function keys(array $filter) { return []; }

    public static function make() {}

    public function many(int ...$ids) {}
}

class DiskStore extends Store
{
    public function __construct($cfg = [])
    {
        parent::__construct($cfg);
    }

    /** Disk variant. */
    public function <weak_warning descr="Method 'put' only forwards to its parent; remove it.">put</weak_warning>($k, $v = 0)
    {
        return parent::put($k, $v);
    }

    public function <weak_warning descr="Method 'get' only forwards to its parent; remove it.">get</weak_warning>($key)
    {
        // just delegate
        parent::get($key);
    }

    public function drop($key)
    {
        parent::drop(trim($key));
    }

    /**
     * @internal
     */
    public function tag($key)
    {
        parent::tag($key);
    }

    /**
     * @api
     */
    public function <weak_warning descr="Method 'flush' only forwards to its parent; remove it.">flush</weak_warning>($all)
    {
        parent::flush($all);
    }

    public function label($text = __CLASS__)
    {
        parent::label($text);
    }

    public function size(): ?int
    {
        return parent::size();
    }

    public function keys($filter)
    {
        return parent::keys($filter);
    }

    #[Cached]
    public static function make()
    {
        parent::make();
    }

    public function <weak_warning descr="Method 'many' only forwards to its parent; remove it.">many</weak_warning>(int ...$ids)
    {
        parent::many($ids);
    }
}
