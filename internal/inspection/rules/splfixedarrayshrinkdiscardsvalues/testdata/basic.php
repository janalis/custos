<?php
$a = SplFixedArray::fromArray(["east", "west"]); <warning descr="Preserve populated entries before shrinking.">$a->setSize(1)</warning>;
