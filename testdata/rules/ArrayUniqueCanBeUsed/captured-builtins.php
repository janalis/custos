<?php
namespace Stats {
    function array_unique(array $a) { return $a; }
    function array_values(array $a) { return $a; }

    function distinct(array $votes) {
        return <weak_warning descr="Use 'array_values(array_unique($votes))' instead (array_unique() is fast since PHP 7.2).">array_keys(array_count_values($votes))</weak_warning>;
    }
}
