<?php
// Option names are matched case-insensitively.
setcookie("auth_token", $token, ["secure" => true, "HttpOnly" => <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>]);
setrawcookie("auth_token", $token, ["HTTPONLY" => <warning descr="Enable HttpOnly for the authentication cookie.">0</warning>, 5 => "ignored", "x"]);
// The element that comes last in array order decides.
setcookie("auth_token", $token, ["httponly" => true, "HttpOnly" => <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>]);
setcookie("auth_token", $token, ["HttpOnly" => true, "httponly" => <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>, "HttpOnly" => true]);
// The legacy httponly argument decides whatever the expiry expression is.
setcookie("auth_token", $token, time() + 3600, "/", "", true, <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
setcookie("auth_token", $token, mktime(0, 0, 0, 1, 1, 2030), "/", "", true, <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
setcookie("auth_token", $token, strtotime("+1 day"), "/", "example.com", true, <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
setcookie("auth_token", $token, $expires, "/", "", true, <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
setcookie("auth_token", $token, 0, "/", "", true, <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
setcookie("auth_token", $token, httponly: <warning descr="Enable HttpOnly for the authentication cookie.">false</warning>);
