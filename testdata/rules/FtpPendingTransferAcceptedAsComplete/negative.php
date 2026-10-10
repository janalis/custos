<?php
if (ftp_nb_get($ftp, $local, $remote, FTP_BINARY) === FTP_FINISHED) { return true; }
