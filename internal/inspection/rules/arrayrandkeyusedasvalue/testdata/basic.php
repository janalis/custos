<?php
function color($v){switch($v){case "red":break;case "blue":break;}} $a=["red","blue"]; color(<warning descr="Read the array element at the returned random key.">array_rand($a)</warning>);
