<?php
function download($ftp, $local, $remote): bool {
    if ($ret = ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) {
        while ($ret === FTP_MOREDATA) { $ret = ftp_nb_continue($ftp); }
        if ($ret !== FTP_FINISHED) { return false; }
        return true;
    }
    return false;
}
function transferFailed($ftp, $local, $remote): bool {
    if (!ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) {
        return true;
    }
    return false;
}
function uploadFailed($ftp, $local, $remote): bool {
    if (ftp_nb_put($ftp, $remote, $local, FTP_BINARY) == false) {
        return true;
    }
    return false;
}
function poll($ftp, $local, $remote): bool {
    if (ftp_nb_get($ftp, $local, $remote, FTP_BINARY)) {
        while (ftp_nb_continue($ftp) !== FTP_FINISHED) { usleep(1000); }
        return true;
    }
    return false;
}
