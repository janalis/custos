<?php
$d=new DOMDocument();$d->loadXML('<root xmlns="urn:demo"><item/></root>');$x=new DOMXPath($d);<warning descr="Register a namespace prefix for the XPath query.">$x->query("//item")</warning>;
