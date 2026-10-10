<?php
// @noinspection IntlTimezoneOffsetMillisecondsUsedAsSeconds
$z=IntlTimeZone::createTimeZone("Europe/Paris"); $z->getOffset($instant,false,$raw,$dst); $d=new DateTime(); $d->modify(($raw+$dst)." seconds");
