<?php
function reviewShapes(ZipArchive $z, $x, $log) {
    if ($x) {
    } elseif (<warning descr="Distinguish a missing ZIP entry from index zero.">!$z->locateName('a')</warning>) {
        echo "missing";
    } elseif (<warning descr="Distinguish a missing ZIP entry from index zero.">$z->locateName('b')</warning>) {
        echo "found";
    }
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">$z->locateName('a') == false</warning>) {
        echo "missing";
    }
    if (<warning descr="Distinguish a missing ZIP entry from index zero.">$z->locateName('a') /* keep */ != false</warning>) {
        echo "found";
    }
    $missing = <warning descr="Distinguish a missing ZIP entry from index zero.">false == ($i = $z->locateName('a'))</warning>;
    $found = <warning descr="Distinguish a missing ZIP entry from index zero.">$z->locateName('a') <> (false)</warning>;
    do {
    } while (<warning descr="Distinguish a missing ZIP entry from index zero.">!($i = $z->locateName('a'))</warning>);
    while (<warning descr="Distinguish a missing ZIP entry from index zero.">($z->locateName('a'))</warning>) {
        break;
    }
    return [$missing, $found, $log];
}
