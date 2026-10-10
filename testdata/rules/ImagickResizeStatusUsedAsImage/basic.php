<?php
$i=new Imagick();$r=$i->resizeImage(80,80,1,1);<warning descr="Use the mutated Imagick object after resizing.">$r->getImageBlob()</warning>;$r=$i->resizeImage(80,80,1,1);echo <warning descr="Use the mutated Imagick object after resizing.">$r->name</warning>;
