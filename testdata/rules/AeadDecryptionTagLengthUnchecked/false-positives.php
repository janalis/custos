<?php
function constantFirst($data, $key, $iv) {
    $tag = $_POST["tag"];
    if (16 !== strlen($tag)) {
        return null;
    }
    return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $tag);
}
function measuredInPlace($data, $key, $iv) {
    if (strlen($_COOKIE["mac"]) !== 16) {
        return null;
    }
    return openssl_decrypt($data, "aes-128-gcm", $key, OPENSSL_RAW_DATA, $iv, $_COOKIE["mac"]);
}
function conditional($data, $key, $iv) {
    $tag = $_GET["t"];
    return strlen($tag) === 16 ? openssl_decrypt($data, "aes-192-gcm", $key, OPENSSL_RAW_DATA, $iv, $tag) : false;
}
function conditionalElse($data, $key, $iv) {
    return 16 !== strlen($_GET["t"]) ? false : openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $_GET["t"]);
}
function branches($data, $key, $iv) {
    if (16 === strlen($_POST["tag"])) {
        return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $_POST["tag"]);
    }
    if (16 !== strlen($_POST["mac"])) {
        return null;
    } else {
        return openssl_decrypt($data, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $_POST["mac"]);
    }
}
function repeated($chunks, $key, $iv) {
    $tag = $_POST["tag"];
    if (16 !== strlen($tag)) {
        return;
    }
    foreach ($chunks as $chunk) {
        openssl_decrypt($chunk, "aes-256-gcm", $key, OPENSSL_RAW_DATA, $iv, $tag);
    }
}
function guardedEarlier($data, $key, $iv) {
    $tag = $_POST["tag"];
    if (16 !== strlen($tag)) {
        return null;
    }
    $mode = OPENSSL_RAW_DATA;
    $aad = '';
    return openssl_decrypt($data, "aes-256-gcm", $key, $mode, $iv, $tag, $aad);
}
