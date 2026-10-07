<?php
function isList($o) {
    return <weak_warning descr="Use 'is_countable($o-&gt;items())' instead of the is_array()/instanceof Countable pair.">is_array($o->Items())</weak_warning> || $o->items() instanceof \Countable;
}
