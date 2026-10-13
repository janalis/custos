<?php
$d=new DOMDocument(); $d->loadXML('<box><part/></box>'); $d->documentElement->appendChild($d->createTextNode('added'));
