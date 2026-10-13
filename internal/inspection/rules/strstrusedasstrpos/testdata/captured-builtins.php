<?php
namespace Text {
    function strpos($h, $n) { return 0; }
    use function Other\stripos;

    if (<warning descr="Use '\strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) {}
    if (<warning descr="Use '\stripos($text, $word) === false' instead; it avoids building a substring.">!stristr($text, $word)</warning>) {}
}

namespace Plain {
    if (<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) {}
}
