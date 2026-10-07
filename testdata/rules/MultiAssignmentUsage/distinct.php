<?php
function distinct(array $r) {
    $a = $r[0];
    <weak_warning descr="Use one destructuring assignment from '$r' instead.">$c = $r[2]</weak_warning>;
    <weak_warning descr="Use one destructuring assignment from '$r' instead.">$d = $r[-2]</weak_warning>;
    <weak_warning descr="Use one destructuring assignment from '$r' instead.">$e = $r[0x3]</weak_warning>;
}
