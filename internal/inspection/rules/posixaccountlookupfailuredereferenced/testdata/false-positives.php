<?php
$a=posix_getpwnam($name);if($a!==false){echo $a["uid"];} $b=[];echo $b["uid"];function dead(){return;echo posix_getpwnam("x")["uid"];}

strlen("unrelated");

function retried_home($login, $suffix) {
    $attempts = 0;
    again:
    try {
        $account = posix_getpwnam($login);
        if (!is_array($account)) {
            return null;
        }
        return $account['dir'] . '/' . $suffix . $account[$suffix];
    } catch (RuntimeException $failure) {
        if (++$attempts < 3) {
            goto again;
        }
        throw $failure;
    }
}
