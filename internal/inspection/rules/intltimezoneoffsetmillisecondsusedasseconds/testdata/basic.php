<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); <warning descr="Convert timezone offsets to seconds.">$d->modify(($raw+$dst)." seconds")</warning>;
