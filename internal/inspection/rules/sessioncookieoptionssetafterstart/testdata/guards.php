<?php
function reopened() {
    session_start(['read_and_close' => true]);
    session_start(['read_and_close' => false, 'cookie_lifetime' => 60]);
    <warning descr="Configure session cookie parameters before starting the session.">session_set_cookie_params(3600)</warning>;
}
function started() {
    if (session_start()) {
        <warning descr="Configure session cookie parameters before starting the session.">session_set_cookie_params(['secure' => true])</warning>;
    }
}
function guarded() {
    if (!session_start()) {
        return;
    }
    <warning descr="Configure session cookie parameters before starting the session.">session_set_cookie_params(lifetime_or_options: 3600)</warning>;
}
function confirmed() {
    session_start();
    if (session_status() === PHP_SESSION_ACTIVE) {
        <warning descr="Configure session cookie parameters before starting the session.">session_set_cookie_params(3600)</warning>;
    }
}
function restarted() {
    if (session_status() !== PHP_SESSION_ACTIVE) {
        session_start();
        <warning descr="Configure session cookie parameters before starting the session.">session_set_cookie_params(3600)</warning>;
    }
}
