<?php
// Whether keys are kept decides which entries are populated.
function computedKeys($keep) { $a = SplFixedArray::fromArray([1, 2, 3], $keep); $a->setSize(1); }
function spreadKeys(array $options) { $a = SplFixedArray::fromArray([1, 2, 3], ...$options); $a->setSize(1); }
function literalKeys() { $a = SplFixedArray::fromArray([1, 2, 3], false); <warning descr="Preserve populated entries before shrinking.">$a->setSize(1)</warning>; }
