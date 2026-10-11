<?php

$level = 20;
resetLevel();
echo strlen(gzencode('sample', $level));

function resetLevel()
{
    global $level;
    $level = 5;
}

$other = 20;
$later = include __DIR__ . '/settings.php';
echo strlen(gzencode('sample', $other));
