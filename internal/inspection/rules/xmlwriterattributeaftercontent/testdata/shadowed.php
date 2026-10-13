<?php namespace Independent;
class XMLWriter {}

$w=new XMLWriter();$w->openMemory();$w->startElement("r");$w->text("x");$w->writeAttribute("id","7");
