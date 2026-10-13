<?php
$z=new ZipArchive(); $z->open($p); $h=$z->getStream("a.txt"); echo stream_get_contents($h); $z->close();
