<?php
$i=imagecreate(3,3);$color=imagecolorallocate($i,255,0,0);$red=<warning descr="Look up palette components instead of shifting the index.">($color>>16)&255</warning>;
