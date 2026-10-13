<?php
namespace Storage {
    function is_dir($path) { return false; }
    function sprintf($format, ...$values) { return ''; }

    function make($dir) {
        if (!mkdir($dir) && !\is_dir($dir)) { throw new \RuntimeException(\sprintf('Directory "%s" was not created', $dir)); }
        if (\mkdir($dir) || \is_dir($dir)) {}
    }
}

namespace Imported {
    use function Storage\is_dir;

    function make($dir) {
        if (!mkdir($dir) && !\is_dir($dir)) {}
    }
}
