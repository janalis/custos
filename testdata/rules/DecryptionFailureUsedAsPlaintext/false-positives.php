<?php
// Each sink is reached only after a false decryption result was rejected.
function a1($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false) { return 0; } $n = 1; return strlen($p); }
function a2($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p) { return strlen($p); } return 0; }
function a3($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false) { return 0; } elseif ($k) { return strlen($p); } else { return strlen($p) + 1; } }
function a4($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return 0; } elseif ($p === false) { return 1; } else { return strlen($p) + 1; } }
function a5($c,$k){ $p = openssl_decrypt($c,'x',$k); return $p === false || strlen($p) < 3; }
function a6($c,$k){ $p = openssl_decrypt($c,'x',$k); if (empty($p)) { return 0; } return strlen($p); }
function a7($c,$k){ $p = openssl_decrypt($c,'x',$k); if (false == $p or $k) { return 0; } return strlen($p); }
function a8($c,$k){ if (!($p = openssl_decrypt($c,'x',$k))) { return 0; } return strlen($p); }
function a9($c,$k){ $p = openssl_decrypt($c,'x',$k); return $p ? strlen($p) : 0; }
function a10($c,$k){ $p = openssl_decrypt($c,'x',$k); return !is_string($p) ? 0 : strlen($p); }
function a11($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p != false and $k) { return strlen($p); } return 0; }
function a12($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return 0; } elseif ($p !== false) { return strlen($p); } return 0; }
function viaAssignInCondition(string $c, string $k, string $iv): int
{
    if (($plain = openssl_decrypt($c, 'aes-256-cbc', $k, 0, $iv)) === false) {
        return -1;
    }
    return strlen($plain);
}
function viaAnd(string $c, string $k, string $iv): int
{
    $plain = openssl_decrypt($c, 'aes-256-cbc', $k, 0, $iv);
    if ($plain !== false && strlen($plain) > 3) {
        return 1;
    }
    return -1;
}
function viaTernary(string $c, string $k, string $iv): int
{
    $plain = openssl_decrypt($c, 'aes-256-cbc', $k, 0, $iv);
    return $plain === false ? -1 : strlen($plain);
}
function viaNot(string $c, string $k, string $iv, string $path): int
{
    $plain = openssl_decrypt($c, 'aes-256-cbc', $k, 0, $iv);
    if (!$plain) {
        return -1;
    }
    return (int) file_put_contents($path, $plain);
}
function viaIsString(string $c, string $k, string $iv): int
{
    $plain = openssl_decrypt($c, 'aes-256-cbc', $k, 0, $iv);
    if (is_string($plain)) {
        return strlen($plain);
    }
    return -1;
function a13($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p .= 'x') { return strlen($p); } return 0; }
function a14($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return 0; } elseif (is_string($p)) { return strlen($p); } return 0; }
function a15($c,$k){ $p = openssl_decrypt($c,'x',$k); if (!$p) { echo 0; } else { echo strlen($p); } }
function a16($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return 0; } elseif (!is_string($p)) { return 1; } else { return strlen($p); } }
function a17($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return 0; } elseif (empty($p)) { return 1; } elseif ($c) { return strlen($p); } return 2; }
