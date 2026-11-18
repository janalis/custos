<?php
$a=new DOMDocument(); $b=new DOMDocument(); $a->appendChild($a->importNode($b->createElement('entry'), true));
