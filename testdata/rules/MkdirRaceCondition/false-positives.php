<?php
function prepare($root, $cfg) {
    $made = mkdir("$root/cache");
    $ok = !@mkdir($root);
    mkdir($root . '/tmp') or exit(1);
    mkdir($root . '/tmp') or die('x');
    if (!is_dir($root) && !mkdir($root) && !is_dir($root)) {}
    if (mkdir($root) || is_dir($root)) {}
    if ($cfg) {} elseif (mkdir($root)) {}
    while (mkdir($root)) {}
    $x = (int) mkdir($root);
    foo(mkdir($root));
    echo mkdir($root) ? 1 : 0;
    $cfg->mkdir($root);
    mkdir();
    mkdir($root, 0700, true, null);
    return mkdir($root);
}

class StorageTest
{
    public function run($d) { mkdir($d); }
}
