<?php
session_start(); session_write_close(); <warning descr="Persist session changes before closing the session.">$_SESSION['visits']++</warning>;
