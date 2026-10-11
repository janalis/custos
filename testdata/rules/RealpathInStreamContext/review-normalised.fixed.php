<?php
// realpath() resolves every `.`, `..` and empty segment: a tail that keeps
// one is another string.
$a = realpath(__DIR__ . '/../X/../X/n.php');
$b = realpath(__DIR__ . '/.././X/n.php');
$c = realpath(__DIR__ . '/../X//n.php');
$w = realpath(__DIR__ . '/../X\..\n.php');
$ok = dirname(__DIR__) . '/X/n.php';
// Outside an inclusion an absolute literal is only reported for its `..`.
$lib = realpath('/usr/bin/../lib');
$usr = realpath("/usr/./bin/..");
require '/usr/bin/../lib/boot.php';
// `__FILE__ . '/..'` is not a directory: realpath() fails on it.
$self = realpath(__FILE__ . '/../n.php');
$dir = dirname(dirname(__FILE__)) . '/n.php';
// A comment inside the call would be lost.
$doc = realpath(__DIR__ /* project root */ . '/../X');
// A tighter operator would capture part of the replacement.
$first = realpath(__DIR__ . '/../f')[0];
$neg = -realpath(__DIR__ . '/../g');
$sum = 1 + realpath(__DIR__ . '/../h');
$quiet = @dirname(__DIR__) . '/i';
$tail = dirname(__DIR__) . '/j' . '/x';
$alt = dirname(__DIR__) . '/k' ?? '/tmp';
// A drive-letter literal stays valid in an inclusion.
include 'C:/app/boot.php';
