<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); <warning descr="Remove live DOM matches without skipping shifted nodes.">$n->parentNode->removeChild($n)</warning>; } }
