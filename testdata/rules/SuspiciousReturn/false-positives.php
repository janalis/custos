<?php
function a() {
    try {
        work();
    } finally {
        return 1;
    }
}
function b() {
    try {
        return 1;
    } finally {
        $f = function () { return 2; };
        cleanup();
    }
}
function c() {
    try {
        throw new Exception();
    } catch (Exception $e) {
        return 3;
    }
}
function e() {
    try {
        $f = function () { return 1; };
        $g = fn() => throw new Exception();
        $o = new class { public function m() { return 2; } };
    } finally {
        return 3;
    }
}
