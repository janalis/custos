<?php
function routes($uri, $base, $enc) {
    $r = [];
    $r[] = <weak_warning descr="Use 'strpos($uri, $base) === 0' instead.">substr($uri, 0, strlen($base)) == $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, '/api') !== 0' instead.">'/api' !== substr($uri, 0, 4)</weak_warning>;
    $r[] = <weak_warning descr="Use '\mb_strpos($uri, $base) === 0' instead.">\mb_substr($uri, 0, mb_strlen($base)) === $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_strpos($uri, $base, 0, $enc) !== 0' instead.">mb_substr($uri, 0, mb_strlen($base, $enc), $enc) !== $base</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_strpos($uri, 'abc', 0, 'UTF-8') === 0' instead.">mb_substr($uri, 0, 3, 'UTF-8') === 'abc'</weak_warning>;
    $r[] = <weak_warning descr="Use 'stripos($uri, 'get:') === 0' instead.">mb_strtolower(substr($uri, 0, 4)) === 'get:'</weak_warning>;
    $r[] = <weak_warning descr="Use 'mb_stripos($uri, 'POST') === 0' instead.">strtoupper(mb_substr($uri, 0, mb_strlen('POST'))) == 'POST'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($uri, $base) !== 0' instead.">substr($uri, 0, strlen($base)) <> $base</weak_warning>;
    return $r;
}
