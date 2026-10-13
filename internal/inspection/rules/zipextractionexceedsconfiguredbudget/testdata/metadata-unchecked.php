<?php $z=new ZipArchive();$z->open("x.zip");$z->statIndex(0);<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("out")</warning>;
