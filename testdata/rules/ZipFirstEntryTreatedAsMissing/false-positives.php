<?php
function reviewValid(ZipArchive $z, $log, $i, $x) {
    // A compound assignment does not yield the lookup result.
    if ($log .= $z->locateName('absent.txt')) {
        echo "entered";
    }
    if ($i ??= $z->locateName('a')) {
        echo "entered";
    }
    // Strict comparisons already separate index zero from a miss.
    if ($z->locateName('a') !== false) {
        echo "found";
    }
    if (($i = $z->locateName('a')) === false) {
        echo "missing";
    }
    // Other comparisons and nested operands are not the lookup's truthiness.
    if ($z->locateName('a') == 0) {
        echo "first";
    }
    if ($z->locateName('a') == true) {
        echo "found";
    }
    if ($z->locateName('a') > 2) {
        echo "late";
    }
    if (!!$z->locateName('a')) {
        echo "found";
    }
    if (-$z->locateName('a')) {
        echo "found";
    }
    if ($x && $z->locateName('a') === false) {
        echo "missing";
    }
    $index = $z->locateName('a');
    return $index;
}
