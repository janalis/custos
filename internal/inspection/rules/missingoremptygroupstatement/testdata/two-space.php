<?php
function bench($n) {
  <weak_warning descr="Use a braced block for the body of this construct.">for</weak_warning> ($i = 0; $i < $n; $i++)
    <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($i) echo $i;
}
