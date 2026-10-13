<?php
function returns() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){return;}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){return;}
if($s["size"]<0){return;}
if($s["size"]>20971520-$total){return;}
$total += $s["size"];
}
$z->extractTo("output");
}
function exits() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){exit(1);}
$total += $s["size"];
}
$z->extractTo("output");
}
function nestedLoop() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){foreach([1] as $x){continue;} switch($x){default: break;} throw new RuntimeException();}
$total += $s["size"];
}
$z->extractTo("output");
}
function nestedCallable() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){$skip=function(array $xs){foreach($xs as $x){continue;}}; throw new RuntimeException();}
$total += $s["size"];
}
$z->extractTo("output");
}
function batch(array $archives) {
$z=new ZipArchive();
foreach($archives as $archive) {
$z->open($archive);
if($z->numFiles>1000){continue;}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){continue 2;}
if($s["size"]<0){break 2;}
if($s["size"]>20971520-$total){continue 2;}
$total += $s["size"];
}
$z->extractTo("output");
}
}
