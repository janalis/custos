<?php
// The emptiness test and the read lead to one report.
function messageId() {
    $right = !empty($_SERVER['SERVER_NAME']) ? $_SERVER['SERVER_NAME'] : 'localhost';
    $host = isset($_SERVER['HTTP_HOST']) ? 'x' : 'y';
    return 'abc' . '@' . <error descr="E-mail address built from client-controlled $_SERVER['SERVER_NAME']; validate it against a whitelist.">$right</error> . $host;
}
