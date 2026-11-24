<?php
// A first-class callable opens no stream: the closure opens one per call.
$z = new ZipArchive();
$z->open($p);
$h = $z->getStream(...);
$z->close();
echo stream_get_contents($h);
