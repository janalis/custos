<?php
namespace Stats {
    function count($value) { return 1; }

    function check(array $rows) {
        if (\count($rows) === 0) {}
    }
}

namespace Plain {
    function check(array $rows) {
        if (count($rows) !== 0) {}
    }
}
