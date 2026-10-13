<?php
/** @custos-credential-store */ function saveCredential($hash) {}
/** @param password-string $password */ function bad($password) { <warning descr="Store passwords with a password hashing API.">saveCredential(md5($password))</warning>; }
