<?php
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); <error descr="Inspect the composite reflection type members.">$t->getName()</error>;
