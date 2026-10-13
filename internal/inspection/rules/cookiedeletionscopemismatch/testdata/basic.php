<?php
setcookie("auth","token",["path"=>"/app"]); <warning descr="Delete the cookie using its original path and domain.">setcookie("auth","",["expires"=>1,"path"=>"/"])</warning>;
