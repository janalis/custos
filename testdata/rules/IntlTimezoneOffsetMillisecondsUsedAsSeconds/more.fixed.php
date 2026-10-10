<?php
$z=IntlTimeZone::createTimeZone("UTC");$z->getOffset(0,false,$raw,$dst);$d=new DateTime();$d->modify(($raw / 1000)." seconds");
