<?php
function swap(array $from, array $to, $text) {
    $text = str_replace(['<', ...array_values($from)], ['&lt;', ...array_values($to)], $text);
    return $text;
}

function sheetName(array $bad, $name) {
    $name = str_replace($bad, ' ', $name);
    return str_replace('\'', '', $name);
}

function strip(array $bad, $name) {
    return str_replace([...array_values($bad), '\''], '', $name);
}

function nested(array $bad, $name) {
    return str_replace('\'', '', str_replace($bad, ' ', $name));
}
