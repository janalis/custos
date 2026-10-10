<?php function audit(){return;
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFile("source.txt","note.txt")) { unlink("source.txt"); $z->close(); }
}