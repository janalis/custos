<?php
class Feed extends SimpleXMLElement {}

function check(?SimpleXMLElement $x, ?Feed $f, ?GMP $n, ?DateTime $d, ?MissingThing $m)
{
    // Empty XML elements and GMP zero are empty: no null comparison.
    if (empty($x) || empty($f) || empty($n)) {
        return 1;
    }
    if (<weak_warning descr="Replace with '$d === null'.">empty($d)</weak_warning>) {
        return 2;
    }
    return <weak_warning descr="Replace with '$m === null'.">empty($m)</weak_warning> ? 3 : 0;
}
