<?php
function scan(string $line, string $tag) {
    $hit  = str_contains($line, $tag);
    $hit2 = FALSE !== mb_strpos($line, '#');
    $miss = !\str_contains(trim($line), $tag);
    return [$hit, $hit2, $miss];
}
