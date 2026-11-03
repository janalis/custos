<?php
function f($servers, $n) {
    <weak_warning descr="Use a braced block for the body of this construct.">for</weak_warning> ($i = 0; $i <= $n; $i++) // Try round-robin
    foreach ($servers as $server) {
        echo $server;
    }
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($n) # hash comment
        echo 1;
    <weak_warning descr="Use a braced block for the body of this construct.">while</weak_warning> ($n--) /* block */ echo 2;
}
function g($n) {
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($n)
        echo 3; /* first */ // second
    <weak_warning descr="Use a braced block for the body of this construct.">else</weak_warning>
        echo 4; /* spans
        lines */
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> (!$n) echo 5; /* note */ echo 6;
}
