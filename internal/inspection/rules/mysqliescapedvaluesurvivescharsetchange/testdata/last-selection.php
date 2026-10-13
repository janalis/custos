<?php
function changedTwice(mysqli $db, $input) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('gbk')) {
        if ($db->set_charset('latin1')) {
            <warning descr="Escape values after selecting the connection charset.">$db->query("SELECT '$v'")</warning>;
        }
    }
}
function otherConnectionRestored(mysqli $db, mysqli $other, $input) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('gbk')) {
        if ($other->set_charset('utf8mb4')) {
            <warning descr="Escape values after selecting the connection charset.">$db->query("SELECT '$v'")</warning>;
        }
    }
}
