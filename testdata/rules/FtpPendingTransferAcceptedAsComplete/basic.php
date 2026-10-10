<?php
if (<warning descr="Wait for FTP_FINISHED before reporting success.">ftp_nb_get($ftp, $local, $remote, FTP_BINARY)</warning>) { return true; }
