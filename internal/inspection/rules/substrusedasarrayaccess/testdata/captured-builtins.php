<?php
namespace Buffer {
    function strlen($s) { return 0; }

    function last(string $buf) {
        return <warning descr="Use '($buf[\strlen($buf) - 1] ?? '')' (string offset access) instead.">substr($buf, -1, 1)</warning>;
    }
}

namespace Plain {
    function last(string $buf) {
        return <warning descr="Use '($buf[strlen($buf) - 1] ?? '')' (string offset access) instead.">substr($buf, -1, 1)</warning>;
    }
}
