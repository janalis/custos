<?php
function emptied() {
    $_SESSION = [];
    session_destroy();
    if ($_SESSION['user']) { echo "in"; }
}
function emptiedLongSyntax() {
    $_SESSION = array();
    session_destroy();
    if ($_SESSION['user']) { echo "in"; }
}
function sessionUnset() {
    session_unset();
    session_destroy();
    if ($_SESSION['user']) { echo "in"; }
}
function sameKeyUnset() {
    unset($_SESSION['user']);
    session_destroy();
    if ($_SESSION["user"]) { echo "in"; }
}
function sameNumericKeyUnset() {
    unset($other, $_SESSION[3]);
    session_destroy();
    if ($_SESSION[3]) { echo "in"; }
}
function wholeArrayUnset() {
    unset($_SESSION);
    session_destroy();
    if ($_SESSION['user']) { echo "in"; }
}
