<?php
function f(XMLReader $r) { if($r->moveToAttribute('key')){ $r->moveToElement(); if($r->nodeType === XMLReader::ELEMENT){} } }
function readBeforeAndRestoreAfter(XMLReader $r) { $r->read(); if ($r->moveToFirstAttribute()) { $r->moveToElement(); if ($r->nodeType === XMLReader::ELEMENT) {} } }
function advancedInBranch(XMLReader $r) { $r->read(); if ($r->moveToAttribute('key')) { $r->read(); if ($r->nodeType === XMLReader::ELEMENT) {} } }
function skippedInBranch(XMLReader $r) { $r->moveToElement(); if ($r->moveToAttribute('key')) { $r->next(); if ($r->nodeType === XMLReader::ELEMENT) {} } }
