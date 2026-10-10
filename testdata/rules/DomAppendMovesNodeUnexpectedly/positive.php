<?php
$d=new DOMDocument(); $a=$d->createElement('a'); $b=$d->createElement('b'); $n=$d->createElement('n'); $a->appendChild($n); <warning descr="Clone the node when preserving its existing parent.">$b->appendChild($n)</warning>;
