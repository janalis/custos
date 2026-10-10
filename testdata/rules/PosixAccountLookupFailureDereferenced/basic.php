<?php
$a=posix_getpwnam($name);echo <warning descr="Reject account lookup failure before indexing.">$a["uid"]</warning>;echo <warning descr="Reject account lookup failure before indexing.">posix_getpwuid(5)["name"]</warning>;
