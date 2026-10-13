<?php
unlink($unknown);unlink("other");if($test){unlink("x");}$z=new ZipArchive();if($z->addFile("source","item",0,0,ZipArchive::FL_OPEN_FILE_NOW)){unlink("source");}if($z->addFile("source","item")){unlink("other");}
$z=new ZipArchive();if($z->addFile($unknown,"item")){unlink("x");}
strlen("ok");
