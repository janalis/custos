<?php
$c=new SoapClient($wsdl,["trace"=>true]);$c->__getLastRequest();$d=new SoapClient($wsdl,$opts);$d->__getLastRequest();$e=new SoapClient($wsdl,["trace"=>$unknown]);$e->__getLastRequest();$unknown->__getLastRequest();function dead(){return;$c=new SoapClient("x");$c->__getLastRequest();}

strlen("unrelated");

function unknownSoap(SoapClient $s){$s->__getLastRequest();}

$c=new SoapClient($wsdl,['trace'=>false]);$alias=&$c;unknown($alias);$c->__getLastRequest();
