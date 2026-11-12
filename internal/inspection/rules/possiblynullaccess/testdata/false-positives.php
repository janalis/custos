<?php
namespace Shop;

class Item
{
    public int $qty = 0;
    public ?Item $next = null;

    public function price(): int { return 1; }
    public function date(): ?\DateTimeImmutable { return null; }
    public static function make(): self { return new self(); }
}

class Assert
{
    /** @phpstan-assert !null $value */
    public static function notNull($value): void {}
}

function takeRef(&$value) {}

class Cart
{
    private ?Item $last = null;
    private static ?Item $shared = null;

    public function find(): ?Item { return null; }
    public function refresh(): void {}

    // Narrowed by conditions and early exits.
    public function guarded(?Item $a, ?Item $b, ?Item $c, ?Item $d, ?Item $e, ?Item $f, ?array $rows): void
    {
        if ($a !== null) {
            $a->price();
        }
        if (isset($b)) {
            $b->price();
        }
        if ($c instanceof Item) {
            $c->price();
        }
        if ($d) {
            $d->price();
        }
        if ($e === null) {
            return;
        }
        $e->price();
        foreach ([1, 2] as $i) {
            if (!$f) {
                continue;
            }
            $f->price();
        }
        $rows ??= [];
        echo $rows['a'];
    }

    // Both cannot be null: once one is, the other is not.
    public function either(?Item $a, ?Item $b): void
    {
        if ($a === null && $b === null) {
            return;
        }
        if ($b === null) {
            $a->price();
            return;
        }
        $b->price();
    }

    // Assertions: assert() and a PHPDoc assertion method.
    public function asserted(?Item $a, ?Item $b): void
    {
        assert($a !== null);
        $a->price();
        Assert::notNull($b);
        $b->price();
    }

    // Null-tolerant forms.
    public function tolerant(?Item $item, ?array $rows): void
    {
        $item?->price();
        echo $item?->next->qty;
        echo $item?->date()->format('c');
        echo $item->qty ?? 0;
        echo $item->next->qty ?? 0;
        echo isset($item->qty) ? 1 : 0;
        echo empty($item->next) ? 1 : 0;
        echo $rows['a'] ?? 'none';
        echo isset($rows['a']['b']) ? 1 : 0;
        $this->last->next ??= new Item();
    }

    // Element writes turn null into an array.
    public function writes(?array $rows, ?array $more, ?array $list, ?array $count, ?array $gone, ?array $refs, ?array $each, $any): void
    {
        $any->fill($rows['z']);
        $rows['a'] = 1;
        $more[] = 1;
        $list['a']['b'] = 2;
        [$x, $list['c']] = [1, 2];
        $count['n']++;
        unset($gone['a']);
        $r = &$refs['a'];
        takeRef($refs['b']);
        foreach ([1] as $each['k']) {
        }
        foreach ([[1, 2]] as [$x, $each['l']]) {
        }
    }

    // A property checked, then re-read after a call that may write it.
    public function afterCall(): void
    {
        if ($this->last === null) {
            return;
        }
        $this->refresh();
        $this->last->price();
        if (self::$shared === null) {
            self::$shared = new Item();
        }
        $this->refresh();
        self::$shared->price();
    }

    // A call checked, then made again.
    public function sameCall(Item $item): void
    {
        if ($item->date() !== null) {
            echo $item->date()->format('c');
        }
        if ($item->date() === null) {
            return;
        }
        $date = $item->date();
        echo $date->format('c');
    }

    // Null only from PHPDoc.
    /** @param Item|null $item */
    public function documented($item): void
    {
        $item->price();
        /** @var Item $found */
        $found = $this->find();
        $found->price();
    }

    // Unknown or mixed values.
    public function unknown($item, mixed $value): void
    {
        $item->price();
        $value->price();
    }

    // $this in a closure may be another object.
    public function bound(): \Closure
    {
        return function () {
            return $this->last->price();
        };
    }

    // A property read used as a key or a property name.
    public function key(Item $item, array $rows): void
    {
        echo $rows[$item->qty];
        echo $item->{$item->qty};
    }

    // Class names are never null.
    public function names(): void
    {
        Item::make();
        echo Item::class;
    }
}
