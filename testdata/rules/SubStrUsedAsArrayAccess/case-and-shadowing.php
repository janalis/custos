<?php
namespace App {
    function substr($s, $a, $l = null) { return ''; }

    function first(string $buf) {
        return substr($buf, 0, 1);
    }
}

namespace {
    function head(string $buf) {
        return <warning descr="Use '($buf[0] ?? '')' (string offset access) instead.">SubStr($buf, 0, 1)</warning>;
    }
}
