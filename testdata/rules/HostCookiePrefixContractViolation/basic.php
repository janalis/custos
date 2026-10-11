<?php
<warning descr="Set Secure and Path=/ and omit Domain for __Host- cookies.">setcookie("__Host-auth",$token,["secure"=>true,"path"=>"/app"])</warning>;
<warning descr="Set Secure and Path=/ and omit Domain for __Host- cookies.">setcookie("__Host-mixed", $token, ["Secure" => true, "Path" => "/app"])</warning>;
<warning descr="Set Secure and Path=/ and omit Domain for __Host- cookies.">setcookie("__Host-domain", $token, ["SECURE" => true, "PATH" => "/", "Domain" => "example.test"])</warning>;
