<?php
function transfer($ftp){if($status=ftp_nb_get($ftp,'local','remote',FTP_BINARY)){if($status===FTP_FINISHED)return true;return false;}return false;}
