<?php $z=new ZipArchive();$z->open("x.zip");validateArchive($z);$z->extractTo("out");
