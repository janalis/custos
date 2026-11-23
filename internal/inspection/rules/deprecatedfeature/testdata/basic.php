<?php
namespace Shop;

$id = <warning descr="Replace spl_object_hash(), deprecated as of PHP 8.6: use spl_object_id().">spl_object_hash</warning>($order);
$all = <warning descr="Replace spl_classes(), deprecated as of PHP 8.6: use ReflectionExtension::getClassNames().">spl_classes</warning>();
$key = <warning descr="Replace metaphone(), deprecated as of PHP 8.6.">\metaphone</warning>('Thompson');
$cmp = <warning descr="Replace strcoll(), deprecated as of PHP 8.6: use Collator::compare().">strcoll</warning>('a', 'b');
$parts = <warning descr="Replace mb_split(), deprecated as of PHP 8.6.">mb_split</warning>('\s', $text);
$set = <warning descr="Replace mysqli_get_charset(), deprecated as of PHP 8.6.">mysqli_get_charset</warning>($link);
$stmt = <warning descr="Replace mysqli_stmt_init(), deprecated as of PHP 8.6: pass the query to mysqli_prepare().">mysqli_stmt_init</warning>($link);

function rows(\ArrayIterator $it, \SplFileObject $csv, \mysqli $db, \RecursiveArrayIterator $deep): void
{
    $it-><warning descr="Replace ArrayIterator::asort(), deprecated as of PHP 8.6.">asort</warning>();
    // A partial application refers to it too: calling the closure raises the deprecation.
    $sort = $it-><warning descr="Replace ArrayIterator::asort(), deprecated as of PHP 8.6.">asort</warning>(...);
    $hash = <warning descr="Replace spl_object_hash(), deprecated as of PHP 8.6: use spl_object_id().">spl_object_hash</warning>(?);
    $deep-><warning descr="Replace ArrayIterator::getFlags(), deprecated as of PHP 8.6.">getFlags</warning>();
    $row = $csv-><warning descr="Replace SplFileObject::fgetcsv(), deprecated as of PHP 8.6.">fgetcsv</warning>();
    $db-><warning descr="Replace mysqli::stmt_init(), deprecated as of PHP 8.6.">stmt_init</warning>();
    $s = <warning descr="Pass the query to new mysqli_stmt(): building it without one is deprecated as of PHP 8.6.">new \mysqli_stmt($db)</warning>;
}

sort($names, <warning descr="Replace SORT_LOCALE_STRING, deprecated as of PHP 8.6: sort with Collator.">SORT_LOCALE_STRING</warning>);
define('LIMIT', 10, <warning descr="Remove define()'s third argument: passing it is deprecated as of PHP 8.6.">false</warning>);
define(<warning descr="Rename 'let': this name is deprecated as of PHP 8.6.">'Shop\let'</warning>, 1);
class_alias(Cart::class, <warning descr="Rename 'Is': this name is deprecated as of PHP 8.6.">'Is'</warning>);

function check(string $class, Cart $cart): bool
{
    $q = http_build_query(<warning descr="Pass an array to http_build_query(): giving it an object is deprecated as of PHP 8.6; use get_object_vars().">$cart</warning>);
    array_walk(<warning descr="Pass an array to array_walk(): giving it an object is deprecated as of PHP 8.6; use get_object_vars().">$cart</warning>, 'trim');
    return is_a(<warning descr="Pass allow_string: true to is_a(): giving it a class name otherwise is deprecated as of PHP 8.6.">$class</warning>, Cart::class)
        || is_subclass_of(<warning descr="Pass allow_string: true to is_subclass_of(): giving it a class name otherwise is deprecated as of PHP 8.6.">$class</warning>, Cart::class, false);
}

final class Cart
{
    const <warning descr="Rename the class constant 'namespace': this name is deprecated as of PHP 8.6.">namespace</warning> = 'shop';

    public function __construct()
    {
        <warning descr="Remove the value returned from __construct(): returning one is deprecated as of PHP 8.6.">return null;</warning>
    }

    public function __destruct()
    {
        <warning descr="Remove this yield from __destruct(): making it a generator is deprecated as of PHP 8.6.">yield 1</warning>;
    }

    public function __debugInfo(): <warning descr="Declare __debugInfo() as returning array: allowing null is deprecated as of PHP 8.6.">?array</warning>
    {
        return [];
    }

    public function total(array $lines): int
    {
        try {
            return array_sum($lines);
        } finally {
            <warning descr="Move this return out of the finally block: returning there is deprecated as of PHP 8.6.">return 0;</warning>
        }
    }
}

final class Crate
{
    public function __debugInfo(): <warning descr="Declare __debugInfo() as returning array: allowing null is deprecated as of PHP 8.6.">null</warning>
    {
        return null;
    }
}

final class Box
{
    public function __debugInfo(): <warning descr="Declare __debugInfo() as returning array: allowing null is deprecated as of PHP 8.6.">array|null</warning>
    {
        return [];
    }
}

interface <warning descr="Rename 'Let': this name is deprecated as of PHP 8.6.">Let</warning> {}
function <warning descr="Rename 'readonly': this name is deprecated as of PHP 8.6.">readonly</warning>() {}
const <warning descr="Rename '_': this name is deprecated as of PHP 8.6.">_</warning> = 1;
