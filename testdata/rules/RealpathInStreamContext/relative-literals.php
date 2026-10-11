<?php
include <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('lib/boot.php')</warning>;
require_once <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath("./bootstrap.php")</warning>;
$up   = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('../../storage')</warning>;
$win  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('C:\app\..\data')</warning>;
$phar = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('phar://app.phar/../x')</warning>;
include <warning descr="Use ''/srv/www/index.php'' instead: realpath() fails inside stream wrappers.">realpath('/srv/www/index.php')</warning>;
