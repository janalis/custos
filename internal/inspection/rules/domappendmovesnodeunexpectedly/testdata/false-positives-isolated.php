<?php
$d=new DOMDocument(); $a=$d->createElement('a'); $b=$d->createElement('b'); $n=$d->createElement('n'); $a->appendChild($n); $b->appendChild($n->cloneNode(true));
