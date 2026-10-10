<?php
/** @custos-protected */ function protectedAction() {}
if ($_SERVER['HTTP_X_FORWARDED_FOR'] === '127.0.0.1') { echo 'log'; }
if ($logging) { echo $_SERVER['HTTP_X_FORWARDED_FOR'] === '127.0.0.1'; protectedAction(); }
if (trustedProxy() && $_SERVER['HTTP_X_FORWARDED_FOR'] === '127.0.0.1') { protectedAction(); }
