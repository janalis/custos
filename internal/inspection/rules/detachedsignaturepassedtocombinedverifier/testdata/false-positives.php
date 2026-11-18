<?php
$sig = sodium_crypto_sign_detached($message, $secret); sodium_crypto_sign_verify_detached($sig, $message, $public);
