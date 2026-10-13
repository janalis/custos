<?php
require __DIR__ . '/lib/boot.php';
include '/srv/app/shared.php';
require 'D:/apps/shared.php';
require 'c://x.php';
include $dynamic;
include '';
include "$base/x.php";
include APP_FILE;
include get_path();

function choose($debug)
{
    $file = $debug ? 'debug.php' : 'prod.php';
    include $file;
}
include 'phar://app.phar/boot.php';
require "file:///srv/app/x.php";
include 'compress.zlib://data.php.gz';
include '\\\\fileserver\\shared\\boot.php';
include "\\\\fileserver\\shared\\boot.php";

function anchored()
{
    $entry = 'main.php';
    $entry = __DIR__ . '/' . $entry;
    return include $entry;
}

function legacyAnchored()
{
    $f = 'a.php';
    $f = dirname(__FILE__) . DIRECTORY_SEPARATOR . $f;
    require_once $f;
}

function resolved()
{
    $entry = 'main.php';
    $entry = realpath($entry);
    include $entry;
}
