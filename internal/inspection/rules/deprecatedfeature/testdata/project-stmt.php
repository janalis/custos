<?php
// The project declares a class of that name in place of PHP's.
class mysqli_stmt { public function __construct($link) {} }
$s = new mysqli_stmt($db);
