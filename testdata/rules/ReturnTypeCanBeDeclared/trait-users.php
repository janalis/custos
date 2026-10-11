<?php
// A trait method is composed into classes whose subclasses may override
// it: reported, never fixed, unless it is final or private.
trait Describes
{
    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">describe</weak_warning>()
    {
        return 'generic';
    }

    public function <weak_warning descr="Declare ': string' as the return type.">label</weak_warning>()
    {
        return 'label';
    }

    public function <weak_warning descr="Declare ': string' as the return type (update the whole hierarchy with a signature refactoring).">renamed</weak_warning>()
    {
        return 'renamed';
    }

    final public function <weak_warning descr="Declare ': string' as the return type.">sealed</weak_warning>()
    {
        return 'sealed';
    }

    private function <weak_warning descr="Declare ': string' as the return type.">hidden</weak_warning>()
    {
        return 'hidden';
    }

    /** @return int */
    abstract public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">id</weak_warning>();

    /** @return int */
    abstract protected function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">unused</weak_warning>();
}

trait Nested
{
    use Describes;

    public function <weak_warning descr="Declare ': int' as the return type.">nested</weak_warning>()
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
