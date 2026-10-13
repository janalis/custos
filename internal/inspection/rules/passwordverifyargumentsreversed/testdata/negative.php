<?php
$hash = password_hash($password, PASSWORD_DEFAULT); password_verify($password, $hash);
