<?php
$tz=IntlTimeZone::createTimeZone("UTC");$tz->getOffset(0,false,$object->raw,$object->dst);$d=new DateTime();$d->modify($object->raw . " seconds");
