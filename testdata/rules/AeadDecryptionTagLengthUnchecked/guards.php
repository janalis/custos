<?php
function otherElement($data, $key, $iv) {
    if (16 !== strlen($_POST["mac"])) {
        return null;
    }
    return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
}
function otherSource($data, $key, $iv) {
    if (16 !== strlen($_GET["tag"])) {
        return null;
    }
    return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
}
function otherLength($data, $key, $iv) {
    $tag = $_POST["tag"];
    return 12 === strlen($tag) ? openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$tag</warning>) : false;
}
function wrongArm($data, $key, $iv) {
    $tag = $_POST["tag"];
    return strlen($tag) === 16 ? false : openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$tag</warning>);
}
function rewrittenInLoop($chunks, $key, $iv) {
    if (strlen($_POST["tag"]) !== 16) {
        return;
    }
    foreach ($chunks as $chunk) {
        openssl_decrypt($chunk, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
        $_POST["tag"] = $chunk;
    }
}
function reboundAfterGuard($data, $key, $iv) {
    $tag = $_POST["tag"];
    if (16 !== strlen($tag)) {
        return null;
    }
    $tag = $_GET["t"];
    $aad = '';
    return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, <warning descr="Validate the authentication tag length before decrypting.">$tag</warning>, $aad);
}
