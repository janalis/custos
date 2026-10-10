<?php
$i=new Imagick($path);<error descr="Select an index below the image count.">$i->setIteratorIndex($i->getNumberImages())</error>;$j=new Imagick();if($j->newImage(2,2,"red")){<error descr="Select an index below the image count.">$j->setIteratorIndex(1)</error>;}
