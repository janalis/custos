<?php
setcookie('sid', 'abc', ['samesite' => 'None', 'secure' => true]);
setcookie('pref', 'dark', ['samesite' => 'None', 'Secure' => true]);
setcookie('lang', 'fr', ['SameSite' => 'None', 'SECURE' => true, 'path' => '/']);
setcookie('mix', 'x', ['samesite' => 'None', 'secure' => false, 'Secure' => true]);
