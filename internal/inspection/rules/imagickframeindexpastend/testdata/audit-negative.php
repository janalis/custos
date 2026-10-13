<?php
$im=new Imagick(); if (!$im->newImage(1,1,'red')) throw new Exception(); $n=$im->getNumberImages(); if (!$im->newImage(1,1,'blue')) throw new Exception(); var_dump($im->setIteratorIndex($n));
