<?php
namespace Shop;

interface Priced {}
final class Order implements Priced {}
class Line {}
final class Refund {}

function find(): Order { return new Order(); }

function checks(int $qty, string $code, ?Order $maybe, Order $order, Line $line, array $tags, float|int $amount): void
{
    $flags = [
        <warning descr="This check is always true: the value is int.">is_int($qty)</warning>,
        <warning descr="This check is always false: the value is int.">is_string($qty)</warning>,
        !<warning descr="This check is always true: the value is string.">\is_string($code)</warning>,
        <warning descr="This check is always false: the value is string.">is_null(value: $code)</warning>,
        <warning descr="This check is always true: the value is Shop\Order.">$order instanceof Priced</warning>,
        <warning descr="This check is always false: the value is Shop\Order.">$order instanceof Refund</warning>,
        <warning descr="This check is always false: the value is Shop\Line.">$line instanceof Order</warning>,
        <warning descr="This check is always false: the value is array.">$tags instanceof \Countable</warning>,
        <warning descr="This check is always true: the value is array.">is_iterable($tags)</warning>,
        <warning descr="This check is always true: the value is float|int.">is_scalar($amount)</warning>,
        <warning descr="This check is always true: the value is Shop\Order.">is_object($order)</warning>,
        <warning descr="This check is always false: the value is Shop\Order.">find() === null</warning>,
        <warning descr="This check is always true: the value is Shop\Order.">null !== (find())</warning>,
        (int) <warning descr="This check is always true: the value is int.">is_int($qty)</warning>,
        (<warning descr="This check is always true: the value is int.">is_int($qty)</warning>) || $maybe,
        <warning descr="This check is always false: the value is DateTime|false.">is_null(\DateTime::createFromFormat('Y', '2020'))</warning>,
    ];
    if ($maybe === null) {
        return;
    }
    if (<warning descr="This check is always true: the value is Shop\Order.">$maybe !== null</warning>) {
        echo 'set';
    }
}

function chain(array|string $value, int|string $key): string
{
    if (is_string($value)) {
        return $value;
    } elseif (<warning descr="This check is always true: the value is array.">is_array($value)</warning>) {
        return 'list';
    } elseif (is_int($value)) {
        return 'unreachable';
    }
    if (is_int($key)) {
        return 'int';
    } elseif (<warning descr="This check is always false: the value is string.">is_int($key)</warning>) {
        return 'never';
    }
    return 'other';
}
