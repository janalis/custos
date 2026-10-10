<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=$nodes->length-1;$i>=0;$i--){ $n=$nodes->item($i); $n->parentNode->removeChild($n); } }
