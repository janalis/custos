<?php
// PHP 5 reads hexadecimal strings as numbers: '0x1a' == '0x1A' there.
function prefixes($uri) {
    $r = array();
    $r[] = substr($uri, 0, 4) == '0x1A';
    $r[] = strpos($uri, 'tel:') === 0;
    return $r;
}
