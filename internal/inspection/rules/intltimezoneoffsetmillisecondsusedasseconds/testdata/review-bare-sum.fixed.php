<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris");$z->getOffset(0,false,$raw,$dst);$d=new DateTime();$d->modify((($raw + $dst) / 1000) . " seconds");
