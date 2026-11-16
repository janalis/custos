<?php
declare(strict_types=1);

namespace Shop;

function total(array $lines): int { return \count($lines); }

// A write into an offset of null makes it an array.
class Cart
{
    private ?array $lines = null;
    public static ?array $shared = null;

    public function add(string $sku): int
    {
        $this->lines[$sku] = 1;
        return total($this->lines);
    }

    public function share(string $sku): int
    {
        self::$shared[] = $sku;
        return total(self::$shared);
    }

    public function nested(string $sku): int
    {
        $this->lines[$sku]['qty'] = 1;
        return total($this->lines);
    }

    public function maybe(bool $add): int
    {
        if ($add) {
            $this->lines['x'] = 1;
        }
        return total(<warning descr="Check this value for null before passing it to $lines of Shop\total(): it is array|null.">$this->lines</warning>);
    }

    // A call between them does not count, as for a null check.
    public function afterCall(): int
    {
        $this->lines['x'] = 1;
        $this->clear();
        return total($this->lines);
    }

    public function clear(): void { $this->lines = null; }

    // Increments, decrements and destructuring into an offset too.
    public function bump(string $sku): int
    {
        $this->lines[$sku]++;
        return total($this->lines);
    }

    public function drop(string $sku): int
    {
        self::$shared[$sku]--;
        return total(self::$shared);
    }

    public function splitInto(string $path): int
    {
        [$this->lines['head'], $this->lines['tail']] = explode('/', $path, 2);
        return total($this->lines);
    }

    // A destructuring that also sets the whole property resets it.
    public function splitAndReset(): int
    {
        $this->lines['x'] = 1;
        [$this->lines['head'], $this->lines] = [1, null];
        return total(<warning descr="Check this value for null before passing it to $lines of Shop\total(): it is array|null.">$this->lines</warning>);
    }
}

function split(?array $parts): int
{
    [$parts['head'], $parts['tail']] = explode('/', 'a/b', 2);
    return total($parts);
}

function tally(?array $seen, string $key): int
{
    $seen[$key]++;
    return total($seen);
}

function maybeTally(?array $seen, string $key, bool $count): int
{
    if ($count) {
        $seen[$key]--;
    }
    return total(<warning descr="Check this value for null before passing it to $lines of Shop\total(): it is array|null.">$seen</warning>);
}
