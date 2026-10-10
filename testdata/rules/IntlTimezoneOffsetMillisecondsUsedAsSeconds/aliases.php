<?php
use IntlTimeZone as BuiltinClass3;
$z=BuiltinClass3::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); <warning descr="Convert timezone offsets to seconds.">$d->modify(($raw+$dst)." seconds")</warning>;
