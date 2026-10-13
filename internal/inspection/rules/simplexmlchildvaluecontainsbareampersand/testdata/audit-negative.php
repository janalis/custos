<?php
$x=new SimpleXMLElement('<!DOCTYPE r [<!ENTITY custom "value">]><r/>'); $x->addChild('a','&custom;'); echo $x->asXML();
