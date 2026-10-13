<?php
function f(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=$nodes->length-1;$i>=0;$i--){ $n=$nodes->item($i); $n->parentNode->removeChild($n); } }
function corrected(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); $n->parentNode->removeChild($n); $i--; } }
function first(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $n=$nodes->item($i); $n->parentNode->removeChild($n); break; } }
function directCorrected(DOMDocument $d) { $nodes=$d->getElementsByTagName('part'); for($i=0;$i<$nodes->length;$i++){ $nodes->item($i)->parentNode->removeChild($nodes->item($i)); --$i; } }
