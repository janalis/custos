<?php namespace Independent;
class SimpleXMLElement {}

$x=new SimpleXMLElement("<enabled>false</enabled>");$v=(bool)$x;
