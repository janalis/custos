<?php
// Before PHP 8.0 strpos() warns and returns false on an empty needle, so only
// a non-empty literal is rewritten.
function prefixes(string $uri, string $base) {
    $r = [];
    $r[] = <weak_warning descr="Use 'strpos($uri, $base) === 0' instead.">substr($uri, 0, strlen($base)) === $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, '') === 0' instead.">substr($uri, 0, 0) === ''</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, &quot;$base&quot;) === 0' instead.">substr($uri, 0, strlen("$base")) === "$base"</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, '/api') === 0' instead.">substr($uri, 0, 4) === '/api'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, '0x1A') === 0' instead.">substr($uri, 0, 4) == '0x1A'</weak_warning>;
    return $r;
}
