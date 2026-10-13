<?php $h = true; $n = 2; ?>
<div>
<?php if ($h) { // show it
    echo "shown";
} ?>
</div>
<?php for ($i = 0; $i < $n; $i++) {} # loop ?>
<b>item</b>
<?php if ($h) {
    echo "a";
} else { // other
    echo "b";
} ?>
<i>tail</i>
<?php if ($h) {
    echo "c";
} // why ?>
<i>line comment before the tag</i>
<?php if ($h) {
    echo "d";
} /* note */ ?>
<i>block comment before the tag</i>
<?php if ($h) {
    echo "e";
} ?>
<i>no space before the tag</i>
