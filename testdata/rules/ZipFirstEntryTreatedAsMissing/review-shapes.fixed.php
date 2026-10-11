<?php
function reviewShapes(ZipArchive $z, $x, $log) {
    if ($x) {
    } elseif ($z->locateName('a') === false) {
        echo "missing";
    } elseif ($z->locateName('b') !== false) {
        echo "found";
    }
    if ($z->locateName('a') === false) {
        echo "missing";
    }
    if ($z->locateName('a') /* keep */ !== false) {
        echo "found";
    }
    $missing = false === ($i = $z->locateName('a'));
    $found = $z->locateName('a') !== (false);
    do {
    } while (($i = $z->locateName('a')) === false);
    while (($z->locateName('a')) !== false) {
        break;
    }
    return [$missing, $found, $log];
}
