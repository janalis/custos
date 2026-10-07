<?php
function a($q) {
    if ($q) { return true; }
    return false;
}
function b($q) {
    if (!($q > 1)) { return true; }
    return false;
}
function c($q) {
    if ($q > 1) { return (true); }
    return false;
}
function d($q) {
    if ($q > 1) { return 1; }
    return 0;
}
function e($q) {
    if ($q > 1) { return true; } else return false;
}
function f($q) {
    if ($q > 1) { return true; } else if ($q) { return false; }
}
function g($q) {
    if ($q > 1) { return true; }
    log_it();
    return false;
}
function h($q) {
    if ($q > 1) { f(); return true; }
    return false;
}
function i($q) {
    if ($q = 1) { return true; }
    return false;
}
