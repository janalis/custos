<?php
function including()
{
    include 'defaults.php';
    return isset($config);
}

function includingLater()
{
    $ready = isset($config);
    include 'defaults.php';
    return $ready;
}
