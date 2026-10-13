<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); $z->open("bundle.zip",BuiltinClass6::CREATE); if ($z->addFile("source.txt","note.txt")) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; $z->close(); }
