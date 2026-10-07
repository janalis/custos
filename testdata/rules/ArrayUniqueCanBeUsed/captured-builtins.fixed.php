<?php
namespace Stats {
    function array_unique(array $a) { return $a; }
    function array_values(array $a) { return $a; }

    function distinct(array $votes) {
        return \array_values(\array_unique($votes));
    }
}
