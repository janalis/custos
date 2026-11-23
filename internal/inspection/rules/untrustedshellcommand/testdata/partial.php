<?php
// A partial application passes the arguments it binds on every call.
$run = <error descr="Keep untrusted input out of shell command syntax.">exec($_GET['cmd'], ...)</error>;
// The command comes with each call.
$later = exec(?);
