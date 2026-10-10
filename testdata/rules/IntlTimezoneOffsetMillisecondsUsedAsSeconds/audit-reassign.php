<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$raw=5;$d=new DateTime();$d->modify($raw . " seconds");
