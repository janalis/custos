<?php
$z=new ZipArchive();if($z->addFromString("a","one")){echo "added";}else{$z->addFromString("a","fallback");}
