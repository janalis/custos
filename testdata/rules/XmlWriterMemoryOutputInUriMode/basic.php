<?php
$w=new XMLWriter();if($w->openUri($path)){$w->writeElement("r","x");<warning descr="Use URI output for a URI-backed XML writer.">$w->outputMemory()</warning>;}
