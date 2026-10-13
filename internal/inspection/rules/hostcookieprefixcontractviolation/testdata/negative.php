<?php
setcookie("__Host-auth",$token,["secure"=>true,"path"=>"/"]);
setcookie("__Host-session", "token", ["Secure" => true, "Path" => "/", "HttpOnly" => true, "SameSite" => "Lax"]);
setcookie("__Host-upper", "token", ["SECURE" => true, "PATH" => "/"]);
setcookie("__Host-twice", "token", ["secure" => true, "path" => "/app", "Path" => "/"]);
