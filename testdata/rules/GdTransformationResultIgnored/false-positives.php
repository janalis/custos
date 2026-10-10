<?php
$i=imagecreatetruecolor(10,10);$s=imagescale($i,5,5);imagepng($s,$path);$u=imagecreatetruecolor(10,10);imagepng($u,$path);function dead(){return;imagepng($i,$path);}

strlen("unrelated");

$i=imagecreate(2,2);imagescale($i,1,1);$alias=&$i;unknown($alias);imagepng($i,$path);
