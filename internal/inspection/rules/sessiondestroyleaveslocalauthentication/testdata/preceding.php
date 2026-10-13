<?php
function otherKeyUnset($k, $data, $i) {
    unset($_SESSION['token']);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function stringVersusNumber($k, $data, $i) {
    unset($_SESSION['01']);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION[01]</warning>) { echo "in"; }
}
function numberVersusString($k, $data, $i) {
    unset($_SESSION[1]);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['1']</warning>) { echo "in"; }
}
function otherNumber($k, $data, $i) {
    unset($_SESSION[1]);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION[2]</warning>) { echo "in"; }
}
function dynamicUnsetKey($k, $data, $i) {
    unset($_SESSION[$k]);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function dynamicReadKey($k, $data, $i) {
    unset($_SESSION['user']);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION[$k]</warning>) { echo "in"; }
}
function otherArrayUnset($k, $data, $i) {
    unset($data['user']);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function otherVariableUnset($k, $data, $i) {
    unset($data);
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function nonEmptyArray($k, $data, $i) {
    $_SESSION = ['user' => 1];
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function otherTarget($k, $data, $i) {
    $data = [];
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function compoundAssign($k, $data, $i) {
    $_SESSION += [];
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function referenceAssign($k, $data, $i) {
    $_SESSION = &$data;
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function scalarAssign($k, $data, $i) {
    $x = 1;
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function otherCall($k, $data, $i) {
    session_write_close();
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function otherExpression($k, $data, $i) {
    $i++;
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
function otherStatement($k, $data, $i) {
    echo 'bye';
    session_destroy();
    if (<warning descr="Clear local session data after destroying the session.">$_SESSION['user']</warning>) { echo "in"; }
}
