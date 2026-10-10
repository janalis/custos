<?php
$s=new DOMDocument(); $s->loadXML('<list><entry/></list>'); $d=new DOMDocument(); $n=$d->importNode($s->documentElement, true);
