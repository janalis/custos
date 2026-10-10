<?php
function f(){$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$raw,$dst);$d=new DateTime();<warning descr="Convert timezone offsets to seconds.">$d->modify($raw . " seconds")</warning>;}
