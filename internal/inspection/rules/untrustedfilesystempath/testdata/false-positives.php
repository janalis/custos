<?php
echo $_GET['file'];

function literalPrefix()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if ($path === false || !str_starts_with($path, '/srv/data/')) {
        exit;
    }
    readfile($path);
}

function comparedLength()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if ($path === false || strncmp($path, '/srv/data/', 10) !== 0) {
        exit;
    }
    readfile($path);
}

function yodaPosition()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if (0 !== strpos($path, '/srv/data/')) {
        return;
    }
    unlink($path);
}

function checkedBase()
{
    $base = realpath(__DIR__ . '/data');
    if ($base === false) {
        throw new RuntimeException('missing');
    }
    $path = realpath($base . '/' . $_GET['file']);
    if ($path === false || strpos($path, $base . DIRECTORY_SEPARATOR) !== 0) {
        throw new RuntimeException('outside');
    }
    return file_get_contents($path);
}

function positiveBranch()
{
    $path = realpath('/srv/data/' . $_GET['file']);
    if ($path !== false && str_starts_with($path, '/srv/data/')) {
        readfile($path);
    }
}

function finiteName()
{
    $name = match ($_GET['doc']) { 'terms' => 'terms.txt', default => 'guide.txt' };
    readfile('/srv/data/' . $name);
}
