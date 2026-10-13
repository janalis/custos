<?php
function c(ZipArchive $z, $x, $p) {
    if ($x) {} elseif (<warning descr="Accept only true from ZipArchive::open.">$z->open($p)</warning>) { echo "ready"; }
    if ($x) {} elseif (!<warning descr="Accept only true from ZipArchive::open.">$z->open($p)</warning>) { echo "failed"; }
    if ($x) {} else if (<warning descr="Accept only true from ZipArchive::open.">$z->open($p)</warning>) { echo "ready"; }
    while (!<warning descr="Accept only true from ZipArchive::open.">$z->open($p)</warning>) { break; }
}
