<?php
// @custos-ignore ReflectionCompositeTypeAssumedNamed
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); $t->getName();
