<?php
$q = new SplPriorityQueue(); $q->insert("first", 2); $q->insert("second", 8); <warning descr="Reverse priorities for smallest-first extraction.">$q->extract()</warning>;
