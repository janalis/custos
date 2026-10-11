<?php
// The node detaches itself before the second append.
function selfRemove() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $a->appendChild($n); $n->remove(); $b->appendChild($n); }
function selfReplace() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $a->appendChild($n); $n->replaceWith($d->createElement('z')); $b->appendChild($n); }
// The node is named in a removal or replacement.
function namedChild() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $a->appendChild($n); $a->removeChild(child: $n); $b->appendChild($n); }
function replaced() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $z = $d->createElement('z'); $a->appendChild($n); $a->replaceChild($z, $n); $b->appendChild($n); }
function inserted() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $z = $d->createElement('z'); $b->appendChild($z); $a->appendChild($n); $b->replaceChild($n, $z); $b->appendChild($n); }
// The first parent drops children that may include the node.
function emptied() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $a->appendChild($n); $a->replaceChildren(); $b->appendChild($n); }
function firstChild() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $a->appendChild($n); $a->removeChild($a->firstChild); $b->appendChild($n); }
function lastChild() { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $z = $d->createElement('z'); $a->appendChild($n); $a->replaceChild($z, $a->lastChild); $b->appendChild($n); }
function unknownNode(DOMNode $n) { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $z = $d->createElement('z'); $a->appendChild($z); $a->appendChild($n); $a->removeChild($z); $b->appendChild($n); }
function adopted(DOMNode $x) { $d = new DOMDocument(); $a = $d->createElement('a'); $b = $d->createElement('b'); $n = $d->createElement('n'); $z = $d->adoptNode($x); $a->appendChild($n); $a->removeChild($z); $b->appendChild($n); }
