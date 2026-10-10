<?php
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<1){throw new RuntimeException();}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
