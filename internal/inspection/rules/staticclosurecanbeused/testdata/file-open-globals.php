<?php
$c = function () { return 1; };
function reach() { return $GLOBALS['c']; }
