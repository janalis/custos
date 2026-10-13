<?php
namespace Feed;

function load(string $body, string $alt) {
    $a = <weak_warning descr="Pass the second argument to state whether JSON decodes to arrays or objects.">json_decode($body)</weak_warning>;
    $b = <weak_warning descr="Pass the second argument to state whether JSON decodes to arrays or objects.">\json_decode(trim($alt))</weak_warning>;
    $c = json_decode($body, false);
    $d = json_decode($body, associative: true);
    $f = <weak_warning descr="Pass the second argument to state whether JSON decodes to arrays or objects.">json_decode($body, depth: 8)</weak_warning>;
    $e = json_decode();
    return [$a, $b, $c, $d, $e, $f];
}
