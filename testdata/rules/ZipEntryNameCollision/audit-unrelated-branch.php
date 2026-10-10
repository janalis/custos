<?php
$z=new ZipArchive();if($flag){$z->addFromString("a","one");}if($z->open("x.zip")){$z->addFromString("a","one");}
