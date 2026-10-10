<?php
$w=new XMLWriter(); $w->openMemory(); $w->startElement('r'); $w->startAttribute('a'); $w->text('v'); $w->endAttribute(); var_dump($w->writeAttribute('b','v')); $w->endElement(); echo $w->outputMemory();
