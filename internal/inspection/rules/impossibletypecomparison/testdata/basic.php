<?php
namespace Shop;

final class Order {}
final class Refund {}
class Line {}
class Fee {}

function total(): int { return 0; }

function compare(int $qty, string $code, Order $order, Refund $refund, Line $line, Fee $fee, array $tags, bool $paid): void
{
    if (<error descr="This comparison is always false: int and string share no value.">$qty === $code</error>) {
        return;
    }
    $changed = <error descr="This comparison is always true: string and int share no value.">$code !== $qty</error>;
    $same = <error descr="This comparison is always false: Shop\Order and Shop\Refund share no value.">$order === $refund</error>;
    $line2 = <error descr="This comparison is always false: Shop\Line and Shop\Fee share no value.">$line === $fee</error>;
    $tagged = <error descr="This comparison is always false: array and string share no value.">$tags === 'sale'</error>;
    $failed = <error descr="This comparison is always false: int and false share no value.">total() === false</error>;
    $number = <error descr="This comparison is always false: int and float share no value.">$qty === 1.0</error>;
    $object = <error descr="This comparison is always false: Shop\Order and int share no value.">(($order)) === $qty</error>;
    $flag = <error descr="This comparison is always false: bool and string share no value.">$paid === 'yes'</error>;
}

function values(int $qty): string
{
    $mode = 'cart';
    if ($qty > 1) {
        $mode = 'bulk';
    }
    if (<error descr="This comparison is always false: 'bulk'|'cart' and 'sale' share no value.">$mode === 'sale'</error>) {
        return 'never';
    }
    $low = 1;
    $high = 2;
    $done = false;
    return <error descr="This comparison is always false: 1 and 2 share no value.">$low === $high</error>
        || <error descr="This comparison is always false: false and true share no value.">$done === true</error> ? 'x' : $mode;
}
