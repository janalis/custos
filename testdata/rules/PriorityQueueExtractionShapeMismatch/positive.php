<?php
$q = new SplPriorityQueue(); $q->insert("task", 5); <warning descr="Match the priority queue extraction mode.">$q->extract()["priority"]</warning>;
