<?php
$root = <warning descr="Use 'dirname(__DIR__)' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . "/../")</warning>;
$etc = <warning descr="Use 'dirname(__DIR__) . &quot;/etc&quot;' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . "/../etc/")</warning> . "/x.xml";
require <warning descr="realpath() fails inside stream wrappers such as phar://; prefer dirname().">realpath('/opt/app/')</warning>;
