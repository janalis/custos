<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$alias=&$raw;$alias=5;$d=new DateTime();$d->modify($raw . " seconds");
