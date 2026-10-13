<?php
function viaStrpos($path) {
    $resolved = realpath($path);
    if ($resolved === false) {
        return false;
    }
    return <warning descr="Require a directory boundary in the containment check.">strpos($resolved, '/srv/files') === 0</warning>;
}

function viaStrposRejection($path) {
    $resolved = realpath($path);
    if ($resolved === false || <warning descr="Require a directory boundary in the containment check.">0 !== strpos($resolved, '/srv/files', 0)</warning>) {
        throw new RuntimeException('outside the upload root');
    }
    return $resolved;
}

function viaStrncmp($path) {
    $root = '/var/app/uploads';
    $resolved = realpath($path);
    if (<warning descr="Require a directory boundary in the containment check.">strncmp($resolved, $root, strlen($root)) == 0</warning>) {
        return true;
    }
    return <warning descr="Require a directory boundary in the containment check.">(strncmp('/var/app/uploads', $resolved, 16)) === 0</warning>;
}

function viaSubstr($path) {
    $resolved = realpath($path);
    if (<warning descr="Require a directory boundary in the containment check.">'/srv/files' != substr($resolved, 0, 10)</warning>) {
        return false;
    }
    return <warning descr="Require a directory boundary in the containment check.">substr($resolved, 0, strlen('/srv/files')) === '/srv/files'</warning>;
}
