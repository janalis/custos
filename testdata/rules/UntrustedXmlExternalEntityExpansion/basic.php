<?php
$d=new DOMDocument(); <warning descr="Disable external entity expansion for request-controlled XML.">$d->loadXML($_POST["xml"],LIBXML_NOENT|LIBXML_DTDLOAD)</warning>;
