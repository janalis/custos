<?php
function sameArgs($a) {
    <error descr="Both compared strings are the same expression; one of them is probably wrong.">strcmp($a->Name(), $a->name())</error>;
    <error descr="Both compared strings are the same expression; one of them is probably wrong.">strcmp(Util::Pick($a), util::pick($a))</error>;
    strcmp($a->name, $a->Name);
}
