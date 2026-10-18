<?php
function documentFromInput(string $xml) {
    $document = new DOMDocument();
    $document->loadXML($xml);
    $document->createElement('item', '&project;');
    <warning descr="Append a text node for literal ampersand text.">$document->createElement('item', '&#0;')</warning>;
}
function documentWithKnownMissingDeclaration() {
    $document = new DOMDocument();
    $document->loadXML('<root/>');
    <warning descr="Append a text node for literal ampersand text.">$document->createElement('item', '&project;')</warning>;
}
function externalDocumentDeclarations() {
    $document = new DOMDocument();
    $document->loadXML('<!DOCTYPE root SYSTEM "local.dtd"><root/>');
    $document->createElement('item', '&project;');
}

function incompleteSubset() {
    $document = new DOMDocument();
    $document->loadXML('<!DOCTYPE x [><x/>');
    $document->createElement('x', '&project;');
}

function multilineSubset() {
    $document = new DOMDocument();
    $document->loadXML('<!DOCTYPE
        root [<!ENTITY project "Example">]><root/>');
    $document->createElement('x', '&project;');
}
function malformedDeclarationHeader() {
    $document = new DOMDocument();
    $document->loadXML('<!DOCTYPEbad><root/>');
    $document->createElement('x', '&project;');
}

function preservedAfterFailedLoad() {
    $document = new DOMDocument();
    $document->loadXML('<!DOCTYPE root [<!ENTITY project "Example">]><root/>');
    $document->loadXML('bad');
    $document->createElement('x', '&project;');
}

final class CustomXmlDocument extends DOMDocument {
    public function loadXML(string $source, int $options = 0): bool { return false; }
}
function customLoader() {
    $document = new CustomXmlDocument();
    $document->loadXML('<root/>');
    $document->createElement('x', '&project;');
}
