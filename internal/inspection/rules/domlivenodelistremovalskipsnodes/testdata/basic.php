<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); <warning descr="Remove live DOM matches without skipping shifted nodes.">$n->parentNode->removeChild($n)</warning>; } }
function direct(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ <warning descr="Remove live DOM matches without skipping shifted nodes.">$nodes->item($i)->parentNode->removeChild($nodes->item($i))</warning>; } }
function trailing(DOMDocument $d, $log) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); <warning descr="Remove live DOM matches without skipping shifted nodes.">$n->parentNode->removeChild($n)</warning>; $log->removed($n); } }
