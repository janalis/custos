<?php
function outerYoda($s, string $t, bool $on)
{
    $a = !('' === (string)$s);
    $b = ('' !== $t) === $on;
    $c = $on && '' === $t;
    return [$a, $b, $c];
}
echo json_encode([outerYoda('abc', '', true), outerYoda('', 'q', false)]), "\n";
