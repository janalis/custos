<?php
$checksum = md5($_POST['payload']); echo $checksum;
/** @custos-credential-store */ function saveCredential($hash) {}
/** @param password-string $password */ function good($password) { saveCredential(password_hash($password, PASSWORD_DEFAULT)); }
