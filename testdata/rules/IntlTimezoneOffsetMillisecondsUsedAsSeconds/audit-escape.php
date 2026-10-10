<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);mutate($raw);$d=new DateTime();$d->modify($raw . " seconds");
