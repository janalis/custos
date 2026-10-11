<?php
// A trait method is composed into classes whose subclasses may override
// it: reported, never fixed, unless it is final or private.
trait Describes
{
    public function describe()
    {
        return 'generic';
    }

    public function label()
    {
        return 'label';
    }

    public function renamed()
    {
        return 'renamed';
    }

    final public function sealed(): string
    {
        return 'sealed';
    }

    private function hidden(): string
    {
        return 'hidden';
    }

    /** @return int */
    abstract public function id();

    /** @return int */
    abstract protected function unused();
}

trait Nested
{
    use Describes;

    public function nested()
    {
        return 1;
    }
}

class Item
{
    use Nested {
        renamed as alias;
        label as protected;
    }

    public function id(): int
    {
        return 1;
    }

    protected function unused(): int
    {
        return 2;
    }
}

class Special extends Item
{
    public function describe(): string
    {
        return 'special';
    }

    public function alias(): string
    {
        return 'alias';
    }
}

class Unrelated
{
    public function label(): string
    {
        return 'unrelated';
    }

    public function nested(): int
    {
        return 2;
    }
}
