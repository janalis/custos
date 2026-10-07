<?php
<error descr="Relative include depends on include_path; anchor it with __DIR__.">require_once 'lib/boot.php'</error>;
<error descr="Relative include depends on include_path; anchor it with __DIR__.">include ("views/header.phtml")</error>;
$cfg = <error descr="Relative include depends on include_path; anchor it with __DIR__.">require('settings.php')</error>;
<error descr="Relative include depends on include_path; anchor it with __DIR__.">include './local.php'</error>;

function plugin()
{
    $entry = 'plugins/main.php';
    <error descr="Relative include depends on include_path; anchor it with __DIR__.">include_once $entry</error>;
}
<error descr="Relative include depends on include_path; anchor it with __DIR__.">include "themes/$theme.php"</error>;
