<?php
function childMovedToAnotherParent() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell'); $o = $d->createElement('other');
    $e->appendChild($c);
    $o->appendChild($c);
    $e->textContent = 'x';
}
function childRemovedItself() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    $c->remove();
    $e->nodeValue = 'x';
}
function childRemovedThroughItsParent() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    $c->parentNode->removeChild($c);
    $e->textContent = 'x';
}
function childMovedWhileRebound() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell'); $o = $d->createElement('other');
    $e->appendChild($c);
    $c = $o->appendChild($c);
    $e->textContent = 'x';
}
function childHandedToUnknownCode($flag) {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    if ($flag) { detach($c); }
    $e->textContent = 'x';
}
function childCalledDynamically($method) {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    $c->$method();
    $e->textContent = 'x';
}
