<?php
// The last matching option in array order enables HttpOnly.
setcookie("auth_token", $token, ["httponly" => false, "HttpOnly" => true]);
setcookie("auth_token", $token, ["httponly" => false, "HTTPONLY" => true, "httponly" => false]);
setcookie("auth_token", $token, ["HttpOnly" => true, "secure" => false]);
// Legacy signature with HttpOnly enabled, unknown or left out.
setcookie("auth_token", $token, time() + 3600, "/", "", true, true);
setcookie("auth_token", $token, time() + 3600, "/", "", true, $httpOnly);
setcookie("auth_token", $token, time() + 3600, "/", "", false);
// Unknown options are not guessed.
setcookie("auth_token", $token, $options);
setcookie("auth_token", $token, [$key => false]);
setcookie("auth_token", $token, [...$options, "httponly" => false]);
// Not a configured authentication cookie.
setcookie("AUTH_TOKEN", $token, ["HttpOnly" => false]);
setcookie("theme", $token, time() + 3600, "/", "", true, false);
