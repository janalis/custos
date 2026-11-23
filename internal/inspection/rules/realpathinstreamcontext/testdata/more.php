<?php
$cb   = realpath(...);
$tpl  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath("$app/../x")</warning>;
$mid  = <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath($app . 'x/..')</warning>;
// A dirname() argument that is not plain hides the base: no fix.
$named = <warning descr="Use 'dirname(dirname(path: __DIR__)) . &quot;/x&quot;' instead: realpath() fails inside stream wrappers.">realpath(dirname(path: __DIR__) . "/../x")</warning>;
