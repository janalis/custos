<?php
$w=new XMLWriter();$w->openMemory();$w->startElement("r");$w->writeAttribute("id","7");$w->text("x");$w->startElement("child");$w->writeAttribute("id","8");$w->endElement();$w->endElement();$w->writeAttribute("x","x");$w->flush();$w->writeAttribute("id","9");$unknown->writeAttribute("x","x");function dead(){return;$w->writeAttribute("x","x");}

strlen("unrelated");

$w=new XMLWriter();$w->endElement();$w->writeAttribute("x","x");$w=new XMLWriter();$w->{"text"}("x");$w->writeAttribute("x","x");

$w=new XMLWriter();$w->startElement('r');$w->text('x');$alias=&$w;unknown($alias);$w->writeAttribute('id','7');
