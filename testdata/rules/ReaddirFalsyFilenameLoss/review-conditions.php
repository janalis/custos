<?php
function review_conditions($h, $skip) {
    if (<warning descr="Compare the directory entry with false.">!!readdir($h)</warning>) { return 1; }
    if (<warning descr="Compare the directory entry with false.">!!!readdir($h)</warning>) { return 2; }
    if (<warning descr="Compare the directory entry with false.">!(readdir($h) == false)</warning>) { return 3; }
    if (<warning descr="Compare the directory entry with false.">(!readdir($h)) === false</warning>) { return 4; }
    if (<warning descr="Compare the directory entry with false.">(readdir($h) == true) !== true</warning>) { return 5; }
    if (<warning descr="Compare the directory entry with false.">\TRUE != readdir($h)</warning>) { return 6; }
    if ($skip) { return 7; } elseif (<warning descr="Compare the directory entry with false.">$e = readdir($h)</warning>) { return $e; }
    if ($skip): elseif (<warning descr="Compare the directory entry with false.">!readdir($h)</warning>): return 8; endif;
    for (; <warning descr="Compare the directory entry with false.">$e = readdir($h)</warning>; ) { echo $e; }
    for ($i = 0; $i++, <warning descr="Compare the directory entry with false.">readdir($h)</warning>; ) { echo $i; }
    do { echo 1; } while (<warning descr="Compare the directory entry with false.">$e = \readdir($h)</warning>);
    return 0;
}
