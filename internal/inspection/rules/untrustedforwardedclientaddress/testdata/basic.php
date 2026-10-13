<?php
/** @custos-protected */ function grantAccess() {}
function bad($allowedIp) { if (<warning descr="Validate the trusted proxy boundary before trusting forwarded addresses.">$_SERVER['HTTP_X_FORWARDED_FOR'] === $allowedIp</warning>) { grantAccess(); } }
