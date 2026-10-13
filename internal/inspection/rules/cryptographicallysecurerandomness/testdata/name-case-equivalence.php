<?php
class Keys { public static $raw; }
function generate() {
    Keys::$raw = openssl_random_pseudo_bytes(32, $strong);
    if (keys::$raw === false || !$strong) {
        throw new \RuntimeException('no entropy');
    }
}
