<?php
namespace Pipeline;

interface Item {}
final class Leaf implements Item {}

interface Stage
{
    public function apply(Item &$item): bool;
}

final class Swap implements Stage
{
    public function apply(Item &$item): bool
    {
        $item = new Leaf();
        return true;
    }
}

final class Logger implements Stage
{
    public function apply(Item &$item): bool
    {
        return $item instanceof Leaf;
    }
}

class Base
{
    public function touch(Item &$item): void {}
}

class Derived extends Base
{
    public function touch(Item &$item): void
    {
        $item = new Leaf();
    }
}

abstract class FromVendor extends \Vendor\Missing\Handler
{
    public function handle(Item &$item): void {}
}

trait Shared
{
    public function share(Item &$item): void {}
}

final class Runner
{
    /** @param Stage[] $stages */
    public function __construct(private array $stages, private Swap $swap) {}

    public function run(Item &$item): void
    {
        foreach ($this->stages as $stage) {
            $stage->apply($item);
        }
    }

    public function direct(Item &$item): void
    {
        $this->swap->apply(item: $item);
    }

    public function passOn(Item &$item): void
    {
        Runner::sink($item);
        unknown_function($item);
    }

    public function unknownMethod(Item &$item): void
    {
        $this->swap->missing($item);
    }

    public function dynamicMethod(Item &$item, string $name): void
    {
        $this->swap->$name($item);
    }

    public function viaCtor(Item &$item): void
    {
        new \Vendor\Thing($item);
    }

    public function ownMethod(<warning descr="Objects are handed over by handle already; drop the '&' before '$item'.">Item &$item</warning>): void
    {
        $this->swap->apply(new Leaf());
        self::sink($item, $item, $item);
        $this->swap->apply(other: $item);
        $name = 'apply';
        $this->swap->$name(new Leaf());
        $cls = Runner::class;
        $cls::sink(new Leaf());
        $copy = new Holder($item);
        \strlen((string) $copy->id);
        \spl_object_id($item);
    }

    public static function sink(Item $first, Item ...$rest): void {}
}

final class Holder
{
    public int $id = 0;

    public function __construct(Item $item) {}
}

interface Tagged {}
class Plain implements Tagged
{
    public function mark(<warning descr="Objects are handed over by handle already; drop the '&' before '$item'.">Item &$item</warning>): void {}
}
class Special extends Plain implements Tagged {}

$anon = new class {
    public function keep(Item &$item): void {}
};

function buffer(array|\ArrayAccess &$context): void
{
    $context['done'] = true;
}

function maybe(<warning descr="Objects are handed over by handle already; drop the '&' before '$item'.">?Item &$item</warning>): void {}
