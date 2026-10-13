<?php
function restored(mysqli $db, $input) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('latin1')) {
        $db->query("SELECT 1");
        if ($db->set_charset('utf8mb4')) {
            $db->query("SELECT '$v'");
        }
    }
}
function restoredUnchecked(mysqli $db, $input) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('gbk')) {
        $db->set_charset('utf8mb4');
        $db->query("SELECT '$v'");
    }
}
function maybeRestored(mysqli $db, $input, $flag) {
    $db->set_charset('utf8mb4');
    $v = $db->real_escape_string($input);
    if ($db->set_charset('gbk')) {
        if ($flag) {
            $db->set_charset('utf8mb4');
        }
        $db->query("SELECT '$v'");
    }
}
function maybeReselectedBeforeEscaping(mysqli $db, $input, $flag) {
    $db->set_charset('utf8mb4');
    if ($flag) {
        $db->set_charset('gbk');
    }
    $v = $db->real_escape_string($input);
    if ($db->set_charset('gbk')) {
        $db->query("SELECT '$v'");
    }
}
