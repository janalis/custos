<?php
$it = new ArrayIterator([7, 11]); iterator_count($it); $it->rewind(); echo $it->current();
