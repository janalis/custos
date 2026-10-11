<?php
function removalOnly(array $items) {
    $q = new SplQueue();
    foreach ($items as $item) {
        <error descr="Check that the collection contains an element.">$q->dequeue()</error>;
    }
}
function rebuilt(array $items) {
    foreach ($items as $item) {
        $q = new SplQueue();
        <error descr="Check that the collection contains an element.">$q->dequeue()</error>;
        $q->enqueue($item);
    }
}
function unbalanced(array $items) {
    $s = new SplStack();
    foreach ($items as $item) {
        $s->push($item);
        $s->pop();
        <error descr="Check that the collection contains an element.">$s->pop()</error>;
    }
}
