<?php
function bootstrap()
{
    $root = <warning descr="Replace dirname(__FILE__) with __DIR__.">dirname(__FILE__)</warning>;
    require <warning descr="Replace dirname(__FILE__) with __DIR__.">\dirname(__FILE__)</warning> . '/config.php';
    echo <warning descr="Replace dirname(__FILE__) with __DIR__.">dirname( __FILE__ )</warning>;
}
