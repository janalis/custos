<?php
$q = new SplPriorityQueue(); $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH); $q->insert("task", 5); $q->extract()["priority"];
