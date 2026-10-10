<?php namespace Independent;
class XMLWriter {}

$w=new XMLWriter();if($w->openUri($path)){$w->writeElement("r","x");$w->outputMemory();}
