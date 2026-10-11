<?php
function acceptedInElse($ftp, $local, $remote): bool {
    if (!<warning descr="Wait for FTP_FINISHED before reporting success.">ftp_nb_get($ftp, $local, $remote, FTP_BINARY)</warning>) {
        return false;
    } else {
        return true;
    }
}
function acceptedWhenNotFalse($ftp, $local, $remote): bool {
    if (($status = <warning descr="Wait for FTP_FINISHED before reporting success.">ftp_nb_put($ftp, $remote, $local, FTP_BINARY)</warning>) != false) {
        echo 'complete';
    }
    return false;
}
