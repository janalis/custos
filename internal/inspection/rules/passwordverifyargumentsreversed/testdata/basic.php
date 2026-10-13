<?php
$hash = password_hash($password, PASSWORD_DEFAULT); <error descr="Pass the password before its stored hash.">password_verify($hash, $password)</error>;
