<?php
namespace App;

<error descr="each() was removed in PHP 8.0; the project targets PHP 8.4.">each</error>($items);
$f = <error descr="create_function() was removed in PHP 8.0; the project targets PHP 8.4.">create_function</error>('$a', 'return $a;');
<error descr="mysqli_bind_param() was removed in PHP 5.4; the project targets PHP 8.4.">mysqli_bind_param</error>($stmt, 's', $v);
<error descr="ldap_sort() was removed in PHP 8.0; the project targets PHP 8.4.">ldap_sort</error>($link, $result, 'cn');
<error descr="mbereg() was removed in PHP 8.0; the project targets PHP 8.4.">mbereg</error>('a', 'b');

// Guarded by name or by version.
if (function_exists('get_magic_quotes_gpc') && get_magic_quotes_gpc()) {
    $v = stripslashes($v);
}
if (PHP_VERSION_ID < 80000) {
    $money = money_format('%i', 1.5);
}

// Functions of an extension PHP moved to PECL, which may still declare them.
mysql_connect();
mcrypt_encrypt('rijndael-128', $key, $data, 'cbc');
imap_open('{host}', 'u', 'p');

// A function with a declaration for the target version.
var_dump($v);
bcdiv('1', '3', 2);
