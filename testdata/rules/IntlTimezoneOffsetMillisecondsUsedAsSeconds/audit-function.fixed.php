<?php
function f(){$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$d=new DateTime();$d->modify(($raw / 1000) . " seconds");}
