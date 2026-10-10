<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$raw/=1000;$d=new DateTime();$d->modify($raw . " seconds");
