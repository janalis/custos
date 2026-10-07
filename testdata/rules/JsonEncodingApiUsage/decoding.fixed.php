<?php
namespace Feed;

function load(string $body, string $alt) {
    $a = json_decode($body, true);
    $b = \json_decode(trim($alt), true);
    $c = json_decode($body, false);
    $d = json_decode($body, associative: true);
    $e = json_decode();
    return [$a, $b, $c, $d, $e];
}
