<?php
function consume(int|string $x) {} $r = new ReflectionFunction("consume"); $t = $r->getParameters()[0]->getType(); foreach ($t->getTypes() as $part) { $part->getName(); }
