<?php
$i=imagecreatetruecolor(10,10);imagescale($i,5,5);<warning descr="Use the image returned by the transformation.">imagepng($i,$path)</warning>;
