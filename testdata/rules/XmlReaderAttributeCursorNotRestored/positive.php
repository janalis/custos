<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ if(<warning descr="Restore the XMLReader element cursor before element operations.">$r->nodeType === XMLReader::ELEMENT</warning>){} } }
