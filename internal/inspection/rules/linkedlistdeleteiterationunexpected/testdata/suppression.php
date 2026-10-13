<?php
// @custos-ignore LinkedListDeleteIterationUnexpected
$l = new SplDoublyLinkedList(); $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE); foreach ($l as $item) {}
