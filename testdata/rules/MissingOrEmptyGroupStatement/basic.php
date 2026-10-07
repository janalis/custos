<?php
function demo($flag, $rows, $n) {
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($flag) log_it();
    <weak_warning descr="Use a braced block for the body of this construct.">elseif</weak_warning> ($n > 3) $n--;
    <weak_warning descr="Use a braced block for the body of this construct.">else</weak_warning> $n = 0;

    if ($flag) { tick(); }
    else <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($n) tock();

    <weak_warning descr="Use a braced block for the body of this construct.">foreach</weak_warning> ($rows as $row) emit($row);
    <weak_warning descr="Use a braced block for the body of this construct.">for</weak_warning> ($i = 0; $i < $n; $i++) ;
    <weak_warning descr="Use a braced block for the body of this construct.">while</weak_warning> (more()) step();
    <weak_warning descr="Use a braced block for the body of this construct.">do</weak_warning> pull(); while (pending());
}
