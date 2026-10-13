<?php
namespace Buffer {
    function strlen($s) { return 0; }

    function last(string $buf) {
        return ($buf[\strlen($buf) - 1] ?? '');
    }
}

namespace Plain {
    function last(string $buf) {
        return ($buf[strlen($buf) - 1] ?? '');
    }
}
