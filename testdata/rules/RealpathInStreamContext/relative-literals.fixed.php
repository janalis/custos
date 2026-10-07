<?php
include realpath('lib/boot.php');
require_once realpath("./bootstrap.php");
$up   = realpath('../../storage');
$win  = 'C:\app\..\data';
$phar = 'phar://app.phar/../x';
include '/srv/www/index.php';
