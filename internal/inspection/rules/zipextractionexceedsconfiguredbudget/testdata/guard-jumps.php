<?php
function skipFalse() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){continue;}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function skipNegative() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){continue;}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function skipBudget() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){continue;}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function stopFalse() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){break;}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function stopNegative() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){break 1;}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function stopBudget() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){break;}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function loggedSkip() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){error_log("too large"); continue;}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function computedLevel() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){break LEVELS;}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function jumpBudget() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){goto extract;}
$total += $s["size"];
}
extract:
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function jumpCount() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){goto extract;}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){throw new RuntimeException();}
$total += $s["size"];
}
extract:
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
function nestedEscape() {
$z=new ZipArchive(); $z->open("bundle.zip");
if($z->numFiles>1000){throw new RuntimeException();}
$total=0;
for($i=0;$i<$z->numFiles;$i++) {
$s=$z->statIndex($i);
if($s===false){throw new RuntimeException();}
if($s["size"]<0){throw new RuntimeException();}
if($s["size"]>20971520-$total){foreach([1] as $x){continue 2;} throw new RuntimeException();}
$total += $s["size"];
}
<warning descr="Validate ZIP extraction against configured limits.">$z->extractTo("output")</warning>;
}
