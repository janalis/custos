<?php
// @custos-ignore PriorityQueueMinOrderAssumption
$q = new SplPriorityQueue(); $q->insert("first", 2); $q->insert("second", 8); $q->extract();
