<?php
function c(ZipArchive $z, $x, $p) {
    if ($x) {} elseif (($z->open($p)===true)) { echo "ready"; }
    if ($x) {} elseif (($z->open($p)!==true)) { echo "failed"; }
    if ($x) {} else if (($z->open($p)===true)) { echo "ready"; }
    while (($z->open($p)!==true)) { break; }
}
