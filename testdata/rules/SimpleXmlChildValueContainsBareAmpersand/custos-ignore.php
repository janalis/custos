<?php
// @custos-ignore SimpleXmlChildValueContainsBareAmpersand

$x=new SimpleXMLElement("<r/>");$x->addChild("label","C & D");
$x=new SimpleXMLElement("<r/>");$x->addChild("label","&bad;");
