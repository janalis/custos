<?php
function loopReassigns($dir, $depth)
{
    $entries = scandir($dir);
    if ($entries === false) {
        return;
    }
    while ($depth--) {
        foreach (<warning descr="Check directory enumeration before iterating.">$entries</warning> as $entry) {
            $dir = $entry;
        }
        $entries = scandir($dir);
    }
}

function testedInsideLoop($dir, $depth)
{
    $entries = scandir($dir);
    while ($depth--) {
        if ($entries === false) {
            return;
        }
        foreach ($entries as $entry) {
            $dir = $entry;
        }
        $entries = scandir($dir);
    }
}
