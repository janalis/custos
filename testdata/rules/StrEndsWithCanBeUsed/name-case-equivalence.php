<?php
function endsWithSuffix($o, $s) {
    return <weak_warning descr="Replace with 'str_ends_with($o-&gt;Path(), $s)'.">substr($o->Path(), -strlen($s)) === $s</weak_warning>;
}
function endsWith($o) {
    return <weak_warning descr="Replace with 'str_ends_with($o-&gt;Path(), $o-&gt;Suffix())'.">substr($o->Path(), -strlen($o->suffix())) === $o->Suffix()</weak_warning>;
}
