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

    }
