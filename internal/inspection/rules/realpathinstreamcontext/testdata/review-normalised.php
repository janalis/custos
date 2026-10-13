<?php
// realpath() resolves every `.`, `..` and empty segment: a tail that keeps
// one is another string.
$a = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(__DIR__ . '/../X/../X/n.php')</warning>;
$b = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(__DIR__ . '/.././X/n.php')</warning>;
$c = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(__DIR__ . '/../X//n.php')</warning>;
$w = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath(__DIR__ . '/../X\..\n.php')</warning>;
$ok = <warning descr="Use 'dirname(__DIR__) . '/X/n.php'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../X/n.php')</warning>;
// Outside an inclusion an absolute literal is only reported for its `..`.
$lib = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('/usr/bin/../lib')</warning>;
$usr = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath("/usr/./bin/..")</warning>;
require <warning descr="Use ''/usr/bin/../lib/boot.php'' instead: realpath() fails inside stream wrappers.">realpath('/usr/bin/../lib/boot.php')</warning>;
// `__FILE__ . '/..'` is not a directory: realpath() fails on it.
$self = <warning descr="Use 'dirname(__FILE__) . '/n.php'' instead: realpath() fails inside stream wrappers.">realpath(__FILE__ . '/../n.php')</warning>;
$dir = <warning descr="Use 'dirname(dirname(__FILE__)) . '/n.php'' instead: realpath() fails inside stream wrappers.">realpath(dirname(__FILE__) . '/../n.php')</warning>;
// A comment inside the call would be lost.
$doc = <warning descr="Use 'dirname(__DIR__) . '/X'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ /* project root */ . '/../X')</warning>;
// A tighter operator would capture part of the replacement.
$first = <warning descr="Use 'dirname(__DIR__) . '/f'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../f')</warning>[0];
$neg = -<warning descr="Use 'dirname(__DIR__) . '/g'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../g')</warning>;
$sum = 1 + <warning descr="Use 'dirname(__DIR__) . '/h'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../h')</warning>;
$quiet = @<warning descr="Use 'dirname(__DIR__) . '/i'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../i')</warning>;
$tail = <warning descr="Use 'dirname(__DIR__) . '/j'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../j')</warning> . '/x';
$alt = <warning descr="Use 'dirname(__DIR__) . '/k'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../k')</warning> ?? '/tmp';
// A drive-letter literal stays valid in an inclusion.
include <warning descr="Use ''C:/app/boot.php'' instead: realpath() fails inside stream wrappers.">realpath('C:/app/boot.php')</warning>;
