<?php
// @custos-ignore ImagickFrameIndexPastEnd

$i=new Imagick($path);$i->setIteratorIndex($i->getNumberImages());$j=new Imagick();if($j->newImage(2,2,"red")){$j->setIteratorIndex(1);}
