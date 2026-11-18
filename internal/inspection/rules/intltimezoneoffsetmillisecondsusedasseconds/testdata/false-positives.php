<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); $d->modify((($raw+$dst)/1000)." seconds");
