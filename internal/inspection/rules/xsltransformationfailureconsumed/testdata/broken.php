<?php
$p=new XSLTProcessor();file_put_contents($out,$p->transformToXML($doc)); $r=$p->transformToXML($doc);echo $r;

$broken = ;
