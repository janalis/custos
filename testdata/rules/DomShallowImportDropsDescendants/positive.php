<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=<warning descr="Import descendants when copying an element subtree.">$d->importNode($s->documentElement)</warning>;
