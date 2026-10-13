<?php namespace Independent;
function imagescale(){}
function imagecrop(){}

$i=imagecreatetruecolor(10,10);imagescale($i,5,5);imagepng($i,$path);
