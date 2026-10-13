<?php
 $x=new SimpleXMLElement("<root><present/></root>");$x->present===null;$x->present==null;$x->present===false;$x=new SimpleXMLElement($unknown);$x->missing===null;$x=new SimpleXMLElement('<root xmlns="urn:x"/>');$x->missing===null;function dead(){return;$x=new SimpleXMLElement("<root/>");$x->missing===null;}

strlen("unrelated");

1===null;$x=new SimpleXMLElement("<r/>");$x->{"missing"}===null;$x->addChild("missing");$x->missing===null;$x=new SimpleXMLElement("broken");$x->missing===null;$x=new SimpleXMLElement("<r/>");$x->missing=1;$x->missing===null;
$unknown->missing===null;
$x=new SimpleXMLElement('<r/>');$alias=&$x;$alias->missing=1;$x->missing===null;
$x=new SimpleXMLElement('<r/>');$x->child[0]='x';$x->missing===null;
$x=new SimpleXMLElement('<r/>');$cb=function()use(&$x){$x->missing=1;};$cb();$x->missing===null;
