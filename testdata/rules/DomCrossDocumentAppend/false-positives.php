<?php
$a = new DOMDocument();
$b = new DOMDocument();

// adoptNode() moves the node itself; its result does not have to be used.
$entry = $b->createElement('entry');
$a->adoptNode($entry);
$a->appendChild($entry);

// The adopted node now accepts children created by its new document.
$entry->appendChild($a->createElement('child'));

// An adoption that may not run leaves the owner unknown.
$maybe = $b->createElement('maybe');
if ($argc > 0) {
    $a->adoptNode($maybe);
}
$a->appendChild($maybe);

// Adopting a parent moves its whole subtree.
$parent = $b->createElement('parent');
$leaf = $b->createElement('leaf');
$parent->appendChild($leaf);
$a->adoptNode($parent);
$a->documentElement->appendChild($leaf);
