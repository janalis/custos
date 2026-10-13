<?php
// A condition that does not reject false on the path to the sink is no guard.
function b1($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false) { echo 1; } return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b2($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false || $k) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b3($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p !== false || $k) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b4($c,$k){ $p = openssl_decrypt($c,'x',$k); return $p === false ? strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>) : 0; }
function b5($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k === false) { return 0; } return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b6($c,$k){ $q = openssl_decrypt($c,'x',$k); $p = openssl_decrypt($c,'y',$k); if ($q === false) { return 0; } return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b7($c,$k){ $p = openssl_decrypt($c,'y',$k); if ($p === false) { return 0; } $p = openssl_decrypt($c,'x',$k); return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b8($c,$k){ $p = openssl_decrypt($c,'x',$k); if (isset($p)) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b9($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p !== null) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b10($c,$k){ $p = openssl_decrypt($c,'x',$k); if (is_string($k)) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b11($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false && $k) { return 0; } return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b12($c,$k){ $p = openssl_decrypt($c,'x',$k); return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>) > 3 && $p !== false; }
function b13($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p === false) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b14($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($k) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } elseif ($p === false) { return 0; } return 0; }
function b15($c,$k){ $p = openssl_decrypt($c,'x',$k); if (-$p) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b16($c,$k){ $p = openssl_decrypt($c,'x',$k); if ($p > false) { return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } return 0; }
function b17($c,$k){ while ($k--) { if ($p === false) { return 0; } $p = openssl_decrypt($c,'x',$k); echo strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); } }
function b18($c,$k){ if (($q = openssl_decrypt($c,'y',$k)) === false) { return 0; } $p = openssl_decrypt($c,'x',$k); return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b19($c,$k){ $p = openssl_decrypt($c,'x',$k); return $p ?: strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>); }
function b20($c,$k){ $p = openssl_decrypt($c,'x',$k); if (strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>) > 2) { return 1; } return 0; }
function b21($c,$k){ $p = openssl_decrypt($c,'x',$k); return strlen(<warning descr="Check decryption success before using the plaintext.">$p</warning>) ? 1 : 0; }
function b23($c,$k){ return strlen(<warning descr="Check decryption success before using the plaintext.">openssl_decrypt($c,'x',$k)</warning>); }
