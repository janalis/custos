<?php
if (<warning descr="Compare the seek result with zero.">fseek($fp, 0)</warning>) { echo "seeked"; }
if ($mode === 1) {
    echo "skip";
} elseif (<warning descr="Compare the seek result with zero.">fseek($fp, 0)</warning>) {
    echo "seeked";
} else if (!<warning descr="Compare the seek result with zero.">fseek($fp, 4)</warning>) {
    echo "failed";
}
for (; <warning descr="Compare the seek result with zero.">fseek($fp, 1, SEEK_CUR)</warning>;) { break; }
for (; $mode, ($r = <warning descr="Compare the seek result with zero.">fseek($fp, 2)</warning>);) { break; }
do { echo 1; } while (<warning descr="Compare the seek result with zero.">fseek($fp, 10)</warning>);
