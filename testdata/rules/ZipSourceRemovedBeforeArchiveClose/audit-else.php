<?php
$z=new ZipArchive();if($z->addFile("a.txt")){echo "added";}else{unlink("a.txt");}
