<?php
// @noinspection ZipEntryNameCollision
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFromString("note.txt","one")) { $z->addFromString("note.txt","two"); }
