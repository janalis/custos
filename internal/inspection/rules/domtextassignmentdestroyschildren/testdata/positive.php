<?php
$d=new DOMDocument(); $d->loadXML('<box><part/></box>'); <warning descr="Preserve child nodes when adding element text.">$d->documentElement->nodeValue='added'</warning>;
