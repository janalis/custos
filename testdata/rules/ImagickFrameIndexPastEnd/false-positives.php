<?php
$i=new Imagick($path);$i->setIteratorIndex(0);$i->setIteratorIndex(-1);$i->setIteratorIndex($other->getNumberImages());$i->getNumberImages();$other->setIteratorIndex(1);function dead(){return;$i=new Imagick();$i->setIteratorIndex($i->getNumberImages());}

strlen("unrelated");

$i=new Imagick();$alias=&$i;unknown($alias);$i->setIteratorIndex($i->getNumberImages());

function reverseWalk(Imagick $im) { $n = $im->getNumberImages(); while ($n-- > 0) { $im->setIteratorIndex($n); } }
function lastFrame(Imagick $im) { $n = $im->getNumberImages(); $n--; $im->setIteratorIndex($n); }
function reboundByLoop(Imagick $im) { $n = $im->getNumberImages(); foreach ([0, 1] as $n) { $im->setIteratorIndex($n); } }
function reboundByList(Imagick $im) { $n = $im->getNumberImages(); list($n) = [0]; $im->setIteratorIndex($n); }
function copiedAfterStep(Imagick $im) { $n = $im->getNumberImages(); $n--; $m = $n; $im->setIteratorIndex($m); }
function reboundByReference(Imagick $im, $zero) { $n = $im->getNumberImages(); $n = &$zero; $im->setIteratorIndex($n); }
