<?php
$w=new XMLWriter();$w->openMemory();$w->startElement("r");$w->text("x");<warning descr="Write attributes before element content.">$w->writeAttribute("id","7")</warning>;
