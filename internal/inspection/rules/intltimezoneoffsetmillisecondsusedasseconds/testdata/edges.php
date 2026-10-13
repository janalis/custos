<?php
$d=new DateTime();$d->modify("now");$d->modify($a." minutes");$d->modify(($a*2)." seconds");$z=IntlTimeZone::createTimeZone("UTC");$z->getOffset(0,false,$raw,$dst);$d->modify(($a+$dst)." seconds");
