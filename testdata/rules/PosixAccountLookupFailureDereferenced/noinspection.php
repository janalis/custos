<?php
// @noinspection PosixAccountLookupFailureDereferenced

$a=posix_getpwnam($name);echo $a["uid"];echo posix_getpwuid(5)["name"];
