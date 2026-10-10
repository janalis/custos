<?php
use ZipArchive as BuiltinClass6;
$z=new BuiltinClass6(); $z->open("bundle.zip",BuiltinClass6::CREATE); if ($z->addFromString("note.txt","one")) { <warning descr="Make ZIP entry replacement explicit.">$z->addFromString("note.txt","two")</warning>; }
