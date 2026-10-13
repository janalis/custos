<?php
$p=new XSLTProcessor();<warning descr="Reject transformation failure before using the XML.">file_put_contents($out,$p->transformToXML($doc))</warning>; $r=$p->transformToXML($doc);<warning descr="Reject transformation failure before using the XML.">echo $r;</warning>
