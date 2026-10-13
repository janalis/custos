<?php
function snapshot() {
    session_start(['read_and_close' => true]);
    session_set_cookie_params(3600);
}
function forwarded(array $options) {
    session_start($options);
    session_set_cookie_params(3600);
}
function dynamicFlag(bool $close) {
    session_start(['read_and_close' => $close]);
    session_set_cookie_params(3600);
}
function failed() {
    if (!session_start()) {
        session_set_cookie_params(3600);
        return;
    }
}
function failedComparison() {
    if (session_start() === false) {
        session_set_cookie_params(['secure' => true]);
    }
}
function failedElse() {
    if (session_start()) {
        $_SESSION['seen'] = true;
    } else {
        session_set_cookie_params(3600);
    }
}
function leftEarly() {
    if (session_start()) {
        return;
    }
    session_set_cookie_params(3600);
}
function statusChecked() {
    session_start();
    if (session_status() !== PHP_SESSION_ACTIVE) {
        session_set_cookie_params(3600);
    }
    if (PHP_SESSION_NONE == session_status()) {
        session_set_cookie_params(3600);
    }
}
function closed() {
    session_start();
    session_write_close();
    session_set_cookie_params(3600);
}
