<?php $h = true; $n = 2; ?>
<div>
<?php <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($h) // show it
    echo "shown" ?>
</div>
<?php <weak_warning descr="Use a braced block for the body of this construct.">for</weak_warning> ($i = 0; $i < $n; $i++) # loop ?>
<b>item</b>
<?php <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($h) echo "a"; <weak_warning descr="Use a braced block for the body of this construct.">else</weak_warning> // other
    echo "b" ?>
<i>tail</i>
<?php <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($h) echo "c" // why ?>
<i>line comment before the tag</i>
<?php <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($h) echo "d" /* note */ ?>
<i>block comment before the tag</i>
<?php <weak_warning descr="Use a braced block for the body of this construct.">if</weak_warning> ($h) echo "e"?>
<i>no space before the tag</i>
