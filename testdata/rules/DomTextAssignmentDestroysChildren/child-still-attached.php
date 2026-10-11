<?php
function childOnlyDecorated() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    $c->setAttribute('id', 'a');
    $c->textContent = 'kept';
    <warning descr="Preserve child nodes when adding element text.">$e->textContent = 'x'</warning>;
}
function childVariableRebound() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    $c = $d->createElement('second');
    $c->remove();
    <warning descr="Preserve child nodes when adding element text.">$e->nodeValue = 'x'</warning>;
}
function childUsedAfterwards() {
    $d = new DOMDocument(); $e = $d->createElement('row'); $c = $d->createElement('cell');
    $e->appendChild($c);
    <warning descr="Preserve child nodes when adding element text.">$e->textContent = 'x'</warning>;
    $c->remove();
}
