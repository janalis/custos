<?php
$z=new ZipArchive(); for ($i=0;$i<$z->numFiles;$i++) { $z->getNameIndex($i); }
