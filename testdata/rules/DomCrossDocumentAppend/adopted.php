<?php
$a = new DOMDocument();
$b = new DOMDocument();
$c = new DOMDocument();

$moved = $b->createElement('moved');
$c->adoptNode($moved);
<error descr="Import the node into the destination document.">$a->appendChild($moved)</error>;

$back = $a->createElement('back');
$b->adoptNode($back);
<error descr="Import the node into the destination document.">$a->insertBefore($back)</error>;
