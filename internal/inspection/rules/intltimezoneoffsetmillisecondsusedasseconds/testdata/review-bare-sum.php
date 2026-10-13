<?php
$z=IntlTimeZone::createTimeZone("Europe/Paris");$z->getOffset(0,false,$raw,$dst);$d=new DateTime();<warning descr="Convert timezone offsets to seconds.">$d->modify($raw + $dst . " seconds")</warning>;
