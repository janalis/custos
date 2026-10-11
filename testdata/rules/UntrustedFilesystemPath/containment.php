<?php
// Canonical paths whose prefix test does not prove containment stay reported.
function shortLength()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if ($path === false || strncmp($path, '/srv/data/', 9) !== 0) {
        exit;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function uncheckedBase()
{
    $base = realpath('/srv/data');
    $path = realpath($base . '/' . $_GET['file']);
    if ($path === false || strpos($path, $base . DIRECTORY_SEPARATOR) !== 0) {
        throw new RuntimeException('outside');
    }
    return <error descr="Validate filesystem containment before using this path.">file_get_contents($path)</error>;
}

function missingSeparator()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (!str_starts_with($path, '/srv/data')) {
        return;
    }
    <error descr="Validate filesystem containment before using this path.">unlink($path)</error>;
}

function notCanonical()
{
    $path = '/srv/data/' . $_GET['file'];
    if (!str_starts_with($path, '/srv/data/')) {
        return;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function reassigned()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (!str_starts_with($path, '/srv/data/')) {
        return;
    }
    $path = $_GET['other'];
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function guardFallsThrough()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (!str_starts_with($path, '/srv/data/')) {
        error_log('outside');
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function elseBranch()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (str_starts_with($path, '/srv/data/')) {
        return;
    } else {
        <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
    }
}

function looseComparison()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (strpos($path, '/srv/data/') != 0) {
        return;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function capturedByReference()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (!str_starts_with($path, '/srv/data/')) {
        return;
    }
    $swap = function () use (&$path) {
        $path = $_GET['other'];
    };
    $swap();
    <error descr="Validate filesystem containment before using this path.">readfile($path)</error>;
}

function requestArm()
{
    $name = match ($_GET['doc']) { 'terms' => 'terms.txt', default => $_GET['doc'] };
    <error descr="Validate filesystem containment before using this path.">readfile('/srv/data/' . $name)</error>;
}

function taintedSuffix()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (!str_starts_with($path, '/srv/data/')) {
        return;
    }
    <error descr="Validate filesystem containment before using this path.">readfile($path . $_GET['suffix'])</error>;
}
