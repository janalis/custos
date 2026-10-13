<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { $z->close(); unlink("source.txt"); }
