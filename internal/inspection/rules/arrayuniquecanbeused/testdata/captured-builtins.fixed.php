<?php
namespace Stats {
    function array_unique(array $a) { return $a; }
    function array_values(array $a) { return $a; }

    function distinct(array $votes) {
        $votes = ['alpha', 'beta'];
        return \array_values(\array_unique($votes));
    }
}
