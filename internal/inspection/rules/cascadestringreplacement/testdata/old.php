<?php
function swap(array $from, array $to, $text) {
    $text = str_replace('<', '&lt;', $text);
    $text = str_replace($from, $to, $text);
    return $text;
}
