<?php
$sig = sodium_crypto_sign_detached($message, $secret); <warning descr="Verify detached signatures with the detached API.">sodium_crypto_sign_open($sig, $public)</warning>;
