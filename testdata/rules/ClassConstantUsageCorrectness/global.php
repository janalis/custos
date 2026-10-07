<?php
use ArrayObject;
use ArrayObject as Bag;

class Local {}

Bag::class;
ArrayObject::class;
<error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">arrayobject</error>::class;
<error descr="Letter case of the class name differs from its declaration; ::class will return the wrong string.">BAG</error>::class;
local::class;
Sub\Thing::class;
