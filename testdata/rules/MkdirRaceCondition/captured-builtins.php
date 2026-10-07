<?php
namespace Storage {
    function is_dir($path) { return false; }
    function sprintf($format, ...$values) { return ''; }

    function make($dir) {
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
        if (<error descr="Re-check with is_dir() after a failed mkdir: 'mkdir($dir) || is_dir(...)'.">\mkdir($dir)</error>) {}
    }
}

namespace Imported {
    use function Storage\is_dir;

    function make($dir) {
        if (<error descr="Re-check with is_dir() after a failed mkdir: '!mkdir($dir) && !is_dir(...)'.">!mkdir($dir)</error>) {}
    }
}
