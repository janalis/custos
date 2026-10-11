<?php
function namedSource($path) { $im = imagecreatetruecolor(8, 8); imagescale(image: $im, width: 2); <warning descr="Use the image returned by the transformation.">imagepng($im, $path)</warning>; }
function namedBoth($path) { $im = imagecreatetruecolor(8, 8); imagescale(width: 2, image: $im); <warning descr="Use the image returned by the transformation.">imagepng(image: $im, file: $path)</warning>; }
function namedCrop($path) { $im = imagecreatetruecolor(8, 8); imagecrop(rectangle: ['x' => 0, 'y' => 0, 'width' => 2, 'height' => 2], image: $im); <warning descr="Use the image returned by the transformation.">imagejpeg($im, $path)</warning>; }
function afterClosure($path) { $im = imagecreatetruecolor(8, 8); $f = function () { return 1; }; imagescale(image: $im, width: 2); <warning descr="Use the image returned by the transformation.">imagegif($im, $path)</warning>; }
