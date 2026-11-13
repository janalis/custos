<?php
namespace Shop;

interface Priced {}
final class Order implements Priced {}
class Line {}
trait Counts
{
    public function counted(): bool
    {
        return $this instanceof Order;
    }
}

class Node
{
    const ROOT = null;
    const MAP = ['a' => 1];

    public function next(): ?Node { return null; }
}

class Cart
{
    // PHPDoc never proves a check redundant.
    /** @param positive-int $qty @param string $code */
    public function documented(int $qty, $code, $any): void
    {
        if ($qty <= 0 || !is_string($code)) {
            throw new \InvalidArgumentException('qty');
        }
        /** @var Order $any */
        if (!$any instanceof Order) {
            throw new \InvalidArgumentException('any');
        }
    }

    // Unknown and mixed values, checks that may go either way.
    public function open(mixed $m, $untyped, ?Order $maybe, Priced $priced, Line $line, iterable $items, callable $cb, object $o): array
    {
        return [
            is_int($m), $untyped === null, $maybe === null, $priced instanceof Order, $line instanceof Priced,
            is_array($items), $items instanceof \Traversable, is_string($cb), is_object($cb), $o instanceof Order,
            is_numeric('1'), is_callable($cb), is_int(...[$m]), is_int(), $m === $untyped, null === null,
        ];
    }

    // A value a loop around the check unsets later.
    public function loops(array $rows): void
    {
        $seen = null;
        foreach ($rows as $row) {
            if ($seen === null) {
                $other = $row;
                foreach ($rows as $r) {
                    $other = $r;
                }
                [$a, $b] = [1, 2];
                list($c) = [3];
                unset($seen);
            }
        }
    }

    // Intersections are not judged.
    public function both(\Countable&\Traversable $x): bool
    {
        return is_object($x);
    }

    // The last branch of a chain names the case left.
    public function last(array|string $value, ?Order $order): string
    {
        if (is_string($value)) {
            return $value;
        } elseif ((is_array($value))) {
            return 'list';
        }
        if ($order === null) {
            return 'none';
        } else if (!is_null($order)) {
            return 'some';
        }
        return '';
    }

    // A value written later in a loop around the check.
    public function walk(Node $first): int
    {
        $count = 0;
        $node = $first;
        while (true) {
            if ($node === null) {
                return $count;
            }
            $count++;
            $node = $node->next();
        }
    }

    // Constants may be configuration; $this in a closure may be another
    // object; assert() states an expectation.
    public function misc(int $qty): \Closure
    {
        assert(is_int($qty));
        $root = Node::ROOT === null || is_int(PHP_INT_SIZE) || is_int(Node::MAP['a']);
        return fn () => $this instanceof Cart && $root;
    }

    // Classes the check cannot judge: self, static, traits, unknown names,
    // expressions.
    public function names(Order $order, string $class): array
    {
        return [
            $this instanceof self, $this instanceof static, $order instanceof Counts,
            $order instanceof Missing, $order instanceof $class,
        ];
    }

    // NotOptimalIfConditions splits if conditions only: here the check is
    // this rule's.
    public function chains(?Order $order, $o): bool
    {
        return ($o instanceof Priced || $o instanceof Order)
            || $order !== null && <warning descr="This check is always true: the value is Shop\Order.">$order instanceof Order</warning>;
    }
}
