<?php
namespace Shop;

interface Priced {}
final class Order implements Priced {}
class Line {}
trait Counts {}

class Cart
{
    const MODE = 'cart';

    private int $count = 0;

    // PHPDoc never proves a comparison impossible.
    /** @param positive-int $qty */
    public function documented(int $qty, $code, $any): bool
    {
        if ($qty <= 0) {
            throw new \InvalidArgumentException('qty');
        }
        /** @var int $any */
        return $qty === 0 || $code === 1 || $any === 'x';
    }

    // Values that may meet, unknown and mixed values.
    public function meet(?Order $maybe, Order $order, Priced $priced, Line $line, mixed $m, $untyped, int|string $key, float $f): array
    {
        return [
            $maybe === $order,
            $priced === $line,
            $m === 1,
            $untyped === 'a',
            $key === 'a',
            $key === 1,
            $f === 1.5,
            $order === $order,
            $key === $untyped,
        ];
    }

    // Ranges and lengths are not exact; values from parameters are not
    // literals.
    public function ranges(int $qty, string $code, $mode = 'a'): array
    {
        if ($qty > 5) {
            return [$qty === 3];
        }
        return [strlen($code) === -1, $mode === 'b', $code === ''];
    }

    // Constants may be configuration; literals compared together are
    // left to other rules; loose comparisons may coerce.
    public function constants(int $qty, string $code): array
    {
        return [PHP_INT_SIZE === 4, self::MODE === 'bulk', static::MODE === 'x', 1 === 2, $qty == '1', $code != 0];
    }

    // Intersections are not judged.
    public function both(\Countable&\Traversable $x): bool
    {
        return $x === 1;
    }

    // A value written later in a loop around the comparison.
    public function loop(array $rows): int
    {
        $step = 'a';
        foreach ($rows as $row) {
            if ($step === 'c') {
                return 1;
            }
            $step = 'b';
        }
        $this->count = 0;
        while ($this->count < 3) {
            if ($this->count === 'x') {
                return 2;
            }
            $this->count++;
        }
        return 0;
    }

    // $this in a closure may be another object; assert() states an
    // expectation.
    public function closures(int $qty): \Closure
    {
        assert($qty !== 'x');
        return function () {
            return $this === 1;
        };
    }

    // NotOptimalIfConditions splits if conditions only: here the
    // comparison is this rule's.
    public function chain($o): bool
    {
        return $o instanceof Order && <error descr="This comparison is always false: Shop\Order and string share no value.">$o === 'x'</error>;
    }
}

// A write into a string offset changes the value, and may lengthen it.
function offsets(): array
{
    $s = 'abc';
    $s[0] = 'x';
    $t = 'abc';
    $t[3] = 'd';
    return [$s === 'xbc', $t === 'abcd', $t !== 'abc'];
}
