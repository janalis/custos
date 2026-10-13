<?php
namespace Html {
    function array_values($a) { return $a; }

    function swap(array $from, array $to, $text) {
        $text = str_replace(['<', ...\array_values($from)], ['&lt;', ...\array_values($to)], $text);
        return $text;
    }
}
