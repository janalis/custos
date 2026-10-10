<?php
$it = new ArrayIterator([7, 11]); iterator_count($it); echo <warning descr="Restore the iterator position after counting.">$it->current()</warning>;
