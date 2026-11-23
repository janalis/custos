<?php
// What PHP deprecated before 8.6, at a PHP 8.5 target.
$day = <warning descr="Replace strftime(), deprecated as of PHP 8.1.">strftime</warning>('%A');
$text = <warning descr="Replace utf8_encode(), deprecated as of PHP 8.2.">utf8_encode</warning>($raw);
$r = <warning descr="Replace lcg_value(), deprecated as of PHP 8.4.">lcg_value</warning>();
<warning descr="Replace curl_close(), deprecated as of PHP 8.5.">curl_close</warning>($ch);
$clean = filter_var($input, <warning descr="Replace FILTER_SANITIZE_STRING, deprecated as of PHP 8.1.">FILTER_SANITIZE_STRING</warning>);
error_reporting(E_ALL & ~<warning descr="Replace E_STRICT, deprecated as of PHP 8.4.">E_STRICT</warning>);
$opt = <warning descr="Replace PDO::MYSQL_ATTR_COMPRESS, deprecated as of PHP 8.5.">PDO::MYSQL_ATTR_COMPRESS</warning>;

function open(ReflectionProperty $p, ReflectionMethod $m): void
{
    $p-><warning descr="Replace ReflectionProperty::setAccessible(), deprecated as of PHP 8.5.">setAccessible</warning>(true);
    $m-><warning descr="Replace ReflectionMethod::setAccessible(), deprecated as of PHP 8.5.">setAccessible</warning>(true);
}

class Moment extends DateTime
{
    public function __wakeup(): void
    {
        parent::<warning descr="Replace DateTime::__wakeup(), deprecated as of PHP 8.5.">__wakeup</warning>();
    }
}

// Not deprecated: the driver classes' own constants, what 8.6 deprecates,
// and a method a class of the project declares.
$ok = Pdo\Mysql::ATTR_COMPRESS;
$hash = spl_object_hash($order);
class Store { public function setAccessible(bool $on): void {} }
function keep(Store $s): void { $s->setAccessible(true); }
// Code that runs only on older PHP versions.
if (PHP_VERSION_ID < 80100) {
    $day = strftime('%A');
}
function legacy(ReflectionProperty $p): void
{
    (\PHP_VERSION_ID < 80100) and $p->setAccessible(true);
}
$flag = version_compare(PHP_VERSION, '8.4', '<') ? E_STRICT : 0;
