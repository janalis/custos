<?php
// An unpacked argument may give any size.
function spreadSize(array $options) { $a = new SplFixedArray(...$options); $a[1] = 1; }
