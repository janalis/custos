<?php
// @noinspection SoapLastRequestWithoutTracing

$c=new SoapClient($wsdl,["trace"=>false]);$c->__getLastRequest();$d=new SoapClient($wsdl,[]);$d->__getLastRequest();$e=new SoapClient($wsdl);$e->__getLastRequest();
