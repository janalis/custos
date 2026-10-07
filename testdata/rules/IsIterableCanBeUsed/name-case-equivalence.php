<?php
function loopable($o) {
    return <weak_warning descr="Use 'is_iterable($o-&gt;rows())' instead of the is_array()/instanceof Traversable pair.">is_array($o->Rows())</weak_warning> || $o->rows() instanceof \Traversable;
}
