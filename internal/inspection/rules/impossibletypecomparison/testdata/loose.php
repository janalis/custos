<?php
namespace Shop;

interface Priced {}
final class Order {}
final class Refund {}
class Cart {}
class Basket extends Cart {}
enum Status { case Open; }
class Stamp extends \DateTime {}
class Feed extends \SimpleXMLElement {}

function ok(): true { return true; }
function maybe(): ?true { return null; }
function none(): null { return null; }

// Loose comparisons PHP decides without looking at the values: objects of
// classes no object belongs to at once, an object against an array, null
// or false, an array against a number or a string, true against false or
// null.
function decided(Order $order, Refund $refund, Cart $cart, Status $status, ?Order $pending, array $list, int $qty, float $rate, string $code): void
{
    $a = <error descr="This comparison is always false: Shop\Order and Shop\Refund never compare equal.">$order == $refund</error>;
    $b = <error descr="This comparison is always true: Shop\Order and Shop\Refund never compare equal.">$order != $refund</error>;
    $c = <error descr="This comparison is always true: Shop\Cart and Shop\Order never compare equal.">$cart <> $order</error>;
    $d = <error descr="This comparison is always false: Shop\Status and Shop\Order never compare equal.">$status == $order</error>;
    $e = <error descr="This comparison is always false: Shop\Order|null and Shop\Refund never compare equal.">$pending == $refund</error>;
    $f = <error descr="This comparison is always false: array and Shop\Cart never compare equal.">$list == $cart</error>;
    $g = <error descr="This comparison is always false: Shop\Order and false never compare equal.">$order == false</error>;
    $h = <error descr="This comparison is always false: array and int never compare equal.">$list == $qty</error>;
    $i = <error descr="This comparison is always false: float and array never compare equal.">$rate == $list</error>;
    $j = <error descr="This comparison is always false: string and array never compare equal.">$code == $list</error>;
    $k = <error descr="This comparison is always false: true and false never compare equal.">ok() == false</error>;
    $l = <error descr="This comparison is always false: true and null never compare equal.">ok() == none()</error>;
}

// Conversions that may make the operands equal stay silent.
function open(Order $order, Cart $cart, Basket $basket, Priced $priced, Stamp $stamp, \DateTimeImmutable $at, Feed $feed, object $any, iterable $items, callable $fn, array $list, ?int $count, bool $flag, int $qty, float $rate, string $code, ?string $name, Status $status, mixed $value, Priced&\Countable $both): void
{
    $a = $order == 1;          // an object converts to 1, with a notice
    $b = $order == 1.0;
    $c = $order == true;       // every plain object is true
    $d = $order == $code;      // __toString in a subclass of a non-final class
    $e = $cart == $basket;     // a Basket is a Cart
    $f = $priced == $order;    // an interface may hold any object
    $g = $stamp == $at;        // DateTime compares across its classes
    $h = $feed == false;       // an empty SimpleXMLElement is false
    $i = $feed == $list;
    $j = $any == $list;
    $k = $items == $qty;       // a Traversable converts to 1
    $l = $fn == $list;         // a callable may be an array
    $m = $list == $count;      // null == []
    $n = $list == $flag;       // [] == false
    $o = $qty == $code;        // 1 == '01'
    $p = $qty == $rate;        // 1 == 1.0
    $q = $count == $code;      // null == ''
    $r = $name == false;       // '0' == false
    $s = $flag == $code;       // true == 'a'
    $t = $status == $code;     // an enum is not its value, but strings stay open
    $u = $value == $order;
    $v = $count == $qty;       // null == 0
    $w = $order == null;       // left to RedundantTypeCheck's null checks
    $x = false == true;        // two literals
    $y = maybe() == false;     // null == false
    $z = $both == $list;       // intersections are not judged
}
