<?php
// @custos-ignore PriorityQueueExtractionShapeMismatch
$q = new SplPriorityQueue(); $q->insert("task", 5); $q->extract()["priority"];
