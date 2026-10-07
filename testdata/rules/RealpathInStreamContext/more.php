<?php
$cb   = realpath(...);
$tpl  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath("$app/../x")</warning>;
$mid  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($app . 'x/..')</warning>;
