<?php
// PHP 5 reads hexadecimal strings as numbers: '0x1a' == '0x1A' there.
function prefixes($uri) {
    $r = array();
    $r[] = <weak_warning descr="Use 'strpos($uri, '0x1A') === 0' instead.">substr($uri, 0, 4) == '0x1A'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, 'tel:') === 0' instead.">substr($uri, 0, 4) == 'tel:'</weak_warning>;
    return $r;
}
