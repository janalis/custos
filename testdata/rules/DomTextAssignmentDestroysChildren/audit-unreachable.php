<?php function audit(){return;
$d=new DOMDocument(); $d->loadXML('<box><part/></box>'); $d->documentElement->nodeValue='added';
}