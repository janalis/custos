<?php namespace Independent;
function imagecolorallocate(){}

$i=imagecreate(3,3);$color=imagecolorallocate($i,255,0,0);$red=($color>>16)&255;
