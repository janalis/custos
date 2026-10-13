<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$f=function()use(&$raw){$raw=5;};$f();$d=new DateTime();$d->modify($raw . " seconds");
