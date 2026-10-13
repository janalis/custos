<?php
function direct(Imagick $im) { $n = $im->getNumberImages(); echo $n; <error descr="Select an index below the image count.">$im->setIteratorIndex($n)</error>; }
function copied(Imagick $im) { $n = $im->getNumberImages(); $m = $n; $n--; <error descr="Select an index below the image count.">$im->setIteratorIndex(($m))</error>; }
function steppedLater(Imagick $im) { $n = $im->getNumberImages(); <error descr="Select an index below the image count.">$im->setIteratorIndex($n)</error>; $n--; }
function reread(Imagick $im) { foreach ([0] as $n) { echo $n; } $n = $im->getNumberImages(); <error descr="Select an index below the image count.">$im->setIteratorIndex($n)</error>; }
function selfCopy(Imagick $im) { $n = $im->getNumberImages(); $n = $n; <error descr="Select an index below the image count.">$im->setIteratorIndex($n)</error>; }
