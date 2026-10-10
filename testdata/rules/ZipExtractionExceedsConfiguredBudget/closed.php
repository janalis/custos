<?php $z=new ZipArchive();$z->open("x.zip");$z->close();$z->extractTo("out");
