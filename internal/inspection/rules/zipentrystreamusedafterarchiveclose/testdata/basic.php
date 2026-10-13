<?php
$z=new ZipArchive(); $z->open($p); $h=$z->getStream("a.txt"); $z->close(); echo <warning descr="Read the entry stream before closing its archive.">stream_get_contents($h)</warning>;
