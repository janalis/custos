<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=$d->importNode($s->documentElement, true);

function prune(DOMElement $element): void
{
    while ($element->firstChild) {
        $element->removeChild($element->firstChild);
    }
}

// The helper may have emptied the element.
function elementPassedToCall(): DOMNode
{
    $source = new DOMDocument();
    $target = new DOMDocument();
    $row = $source->createElement('row');
    $row->appendChild($source->createElement('cell'));
    prune($row);

    return $target->importNode($row);
}

// The only child is removed through the root element.
function rootChildRemoved(): DOMNode
{
    $source = new DOMDocument();
    $source->loadXML('<list><entry/></list>');
    $source->documentElement->removeChild($source->documentElement->firstChild);
    $target = new DOMDocument();

    return $target->importNode($source->documentElement);
}

// The appended child detaches itself.
function childDetached(): DOMNode
{
    $source = new DOMDocument();
    $target = new DOMDocument();
    $row = $source->createElement('row');
    $cell = $source->createElement('cell');
    $row->appendChild($cell);
    $cell->remove();

    return $target->importNode($row);
}

// The root element is swapped for an empty one.
function rootReplaced(): DOMNode
{
    $source = new DOMDocument();
    $source->loadXML('<list><entry/></list>');
    $source->replaceChild($source->createElement('empty'), $source->documentElement);
    $target = new DOMDocument();

    return $target->importNode($source->documentElement);
}

// A later HTML load replaces the proven tree.
function reloadedAsHtml(): DOMNode
{
    $source = new DOMDocument();
    $source->loadXML('<list><entry/></list>');
    $source->loadHTML('<p>x</p>');
    $source->documentElement->replaceChildren();
    $target = new DOMDocument();

    return $target->importNode($source->documentElement);
}
