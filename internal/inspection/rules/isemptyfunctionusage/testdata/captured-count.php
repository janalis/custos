<?php
namespace Stats {
    function count($value) { return 1; }

    function check(array $rows) {
        if (<weak_warning descr="Replace with '\count($rows) === 0'.">empty($rows)</weak_warning>) {}
    }
}

namespace Plain {
    function check(array $rows) {
        if (<weak_warning descr="Replace with 'count($rows) !== 0'.">!empty($rows)</weak_warning>) {}
    }
}
