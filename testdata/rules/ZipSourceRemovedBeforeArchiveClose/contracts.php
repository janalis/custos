<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; $z->close(); }
