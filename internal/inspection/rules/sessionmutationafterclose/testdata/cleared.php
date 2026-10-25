<?php
session_start(); $_SESSION['k'] = 1; session_write_close();
$_SESSION = [];
$_SESSION = array();
$_SESSION = ([]);
<warning descr="Persist session changes before closing the session.">$_SESSION = ['k' => 2]</warning>;
<warning descr="Persist session changes before closing the session.">$_SESSION['k'] = []</warning>;
<warning descr="Persist session changes before closing the session.">$_SESSION += []</warning>;
