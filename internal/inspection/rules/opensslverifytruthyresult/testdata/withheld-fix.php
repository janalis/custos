<?php
function commented($d, $s, $k) {
    if (!(/* keep */ <warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>)) { return false; }
    return true;
}
