<?php
$x=new SimpleXMLElement("<r/>");<warning descr="Escape bare ampersands in XML child values.">$x->addChild("label","C & D")</warning>;
$x=new SimpleXMLElement("<r/>");$x->addChild("label","&bad;");
$x=new SimpleXMLElement("<r/>");<warning descr="Escape bare ampersands in XML child values.">$x->addChild("label","&#xZZ;")</warning>;
