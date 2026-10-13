<?php
include realpath('lib/boot.php');
require_once realpath("./bootstrap.php");
$up   = realpath('../../storage');
$win  = realpath('C:\app\..\data');
$phar = realpath('phar://app.phar/../x');
include '/srv/www/index.php';
