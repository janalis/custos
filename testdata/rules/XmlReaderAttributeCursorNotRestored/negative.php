<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ $r->moveToElement(); if($r->nodeType === XMLReader::ELEMENT){} } }
