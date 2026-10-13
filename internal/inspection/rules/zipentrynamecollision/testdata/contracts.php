<?php
$z=new ZipArchive(); $z->open("bundle.zip",ZipArchive::CREATE); if ($z->addFromString("note.txt","one")) { <warning descr="Make ZIP entry replacement explicit.">$z->addFromString("note.txt","two")</warning>; }
