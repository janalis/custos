<?php
$d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$x=new DOMXPath($d);$x->registerNamespace("d","urn:demo");$x->query("//d:item");
function appended() { $d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$d->documentElement->appendChild($d->createElement('item'));$x=new DOMXPath($d);return $x->query("//item")->length; }
function reloaded() { $d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$d->load(__DIR__.'/plain.xml');$x=new DOMXPath($d);return $x->query("//item")->length; }
function html() { $d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$d->loadHTML('<html><body><item></item></body></html>');$x=new DOMXPath($d);return $x->query("//item")->length; }
