<?php
// @custos-ignore ImagickResizeStatusUsedAsImage

$i=new Imagick();$r=$i->resizeImage(80,80,1,1);$r->getImageBlob();$r=$i->resizeImage(80,80,1,1);echo $r->name;
