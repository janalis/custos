<?php
namespace Text {
    function strpos($h, $n) { return 0; }
    use function Other\stripos;

    if (\strpos($text, $word) !== false) {}
    if (\stripos($text, $word) === false) {}
}

namespace Plain {
    if (strpos($text, $word) !== false) {}
}
