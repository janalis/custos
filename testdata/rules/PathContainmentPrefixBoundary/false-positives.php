<?php
function withBoundary($path) {
    $resolved = realpath($path);
    if ($resolved === false) {
        return false;
    }
    return $resolved === '/srv/files'
        || strpos($resolved, '/srv/files/') === 0
        || strncmp($resolved, '/srv/files/', 11) === 0
        || substr($resolved, 0, 11) === '/srv/files/';
}

function notAPrefixTest($path, $offset, $length, $other) {
    $resolved = realpath($path);
    $found = strpos($resolved, '/srv/files');
    return strpos($resolved, '/srv/files') == 0
        || strpos($resolved, '/srv/files') === 4
        || strpos($resolved, '/srv/files') !== false
        || strpos($resolved, '/srv/files') > 0
        || strpos($resolved, '/srv/files', 1) === 0
        || strpos($resolved, '/srv/files', $offset) === 0
        || strncmp($resolved, '/srv/files', 4) === 0
        || strncmp($resolved, '/srv/files', $length) === 0
        || strncmp($resolved, '/srv/files', strlen('/srv')) === 0
        || strncmp($resolved, '/srv/files', mb_strlen('/srv/files')) === 0
        || strncmp($resolved, '/srv/files', 10) === 1
        || !strncmp($resolved, '/srv/files', 10)
        || substr($resolved, 0) === '/srv/files'
        || substr($resolved, 1, 10) === '/srv/files'
        || substr($resolved, 0, 9) === '/srv/files'
        || substr($resolved, 0, 10) === $other
        || substr($resolved, 0, 10) . '/srv/files'
        || $found;
}

function notCanonical($path) {
    return strpos($path, '/srv/files') === 0
        || strncmp($path, '/srv/files', 10) === 0
        || substr($path, 0, 10) === '/srv/files';
}
