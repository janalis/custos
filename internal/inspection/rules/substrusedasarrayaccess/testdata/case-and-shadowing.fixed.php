<?php
namespace App {
    function substr($s, $a, $l = null) { return ''; }

    function first(string $buf) {
        return substr($buf, 0, 1);
    }
}

namespace {
    function head(string $buf) {
        return ($buf[0] ?? '');
    }
}
