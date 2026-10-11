<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ if(<warning descr="Restore the XMLReader element cursor before element operations.">$r->nodeType === XMLReader::ELEMENT</warning>){} } }
function afterRead(XMLReader $r) { $r->read(); if ($r->moveToFirstAttribute()) { if (<warning descr="Restore the XMLReader element cursor before element operations.">$r->nodeType === XMLReader::ELEMENT</warning>) {} } }
function afterMoveToElement(XMLReader $r) { $r->moveToElement(); if ($r->moveToAttribute('key')) { if (<warning descr="Restore the XMLReader element cursor before element operations.">$r->nodeType === XMLReader::ELEMENT</warning>) {} } }
function afterNext(XMLReader $r) { while ($r->next()) { if ($r->moveToAttribute('key')) { if (<warning descr="Restore the XMLReader element cursor before element operations.">XMLReader::ELEMENT === $r->nodeType</warning>) {} } } }
