<?php
$d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$x=new DOMXPath($d);$x->registerNamespace("d","urn:demo");$x->query("//d:item");
