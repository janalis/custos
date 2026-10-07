<?php
function escape(array $a, array $b, $text) {
    $text = str_replace([...array_values($a), ...array_values(array_keys($b))], [...array_values($b), ...array_values($a)], $text);
    return $text;
}

function twice(array $from1, array $to1, array $from2, array $to2, $text) {
    return str_replace([...array_values($from1), ...array_values($from2)], [...array_values($to1), ...array_values($to2)], $text);
}
