<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=$d->importNode($s->documentElement, true);

function attachedElement(): DOMNode
{
    $source = new DOMDocument();
    $target = new DOMDocument();
    $row = $source->createElement('row');
    $cell = $source->createElement('cell');
    $cell->setAttribute('id', '1');
    $row->appendChild($cell);
    $row->setAttribute('kind', 'data');
    $source->appendChild($row);
    echo $source->saveXML();

    return $target->importNode($row, true);
}
