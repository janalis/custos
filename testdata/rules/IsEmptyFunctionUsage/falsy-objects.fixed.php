<?php
class Feed extends SimpleXMLElement {}

function check(?SimpleXMLElement $x, ?Feed $f, ?GMP $n, ?DateTime $d, ?MissingThing $m)
{
    // Empty XML elements and GMP zero are empty: no null comparison.
    if (empty($x) || empty($f) || empty($n)) {
        return 1;
    }
    if ($d === null) {
        return 2;
    }
    return $m === null ? 3 : 0;
}
