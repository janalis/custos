<?php
function prepare($root, $cfg) {

    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($root) && !is_dir(...)'.">(false === mkdir($root))</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir(&quot;$root/log&quot;, 0700, true) || is_dir(...)'.">(mkdir("$root/log", 0700, true))</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir(dirname($root)) || is_dir(...)'.">(mkdir(dirname($root)))</error>) {}
    if (file_exists($root) || <error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($root, permissions: 0700) || is_dir(...)'.">mkdir($root, permissions: 0700)</error>) {}
    if ($cfg->ok && <error descr="Re-check with is_dir() after a failed mkdir: '!mkdir(strtolower($root)) && !is_dir(...)'.">!mkdir(strtolower($root))</error>) {}
    if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($root) && !is_dir(...)'.">mkdir($root) !== true</error>) {}
}
function tempVar($cfg) {
    if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($cfg->dir(), 0700) || is_dir(...)'.">mkdir($cfg->dir(), 0700)</error>) { return 1; }
    return 0;
}
