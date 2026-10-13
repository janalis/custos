<?php
function rows() { yield 'k' => 2; yield 'k' => 5; } $a = <warning descr="Preserve all yielded values when materializing the iterator.">iterator_to_array(rows())</warning>;
