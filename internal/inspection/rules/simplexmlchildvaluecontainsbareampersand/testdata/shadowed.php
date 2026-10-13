<?php namespace Independent;
class SimpleXMLElement {}

$x=new SimpleXMLElement("<r/>");$x->addChild("label","C & D");
$x=new SimpleXMLElement("<r/>");$x->addChild("label","&bad;");
