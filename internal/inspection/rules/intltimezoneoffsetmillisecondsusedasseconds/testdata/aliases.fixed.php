<?php
use IntlTimeZone as BuiltinClass3;
$z=BuiltinClass3::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); $d->modify((($raw+$dst) / 1000)." seconds");
