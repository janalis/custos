<?php
function conditionalLiteralInsert(bool $more) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    if ($more) { $q->insert('other', 6); $q->count(); }
    return <warning descr="Match the priority queue extraction mode.">$q->extract()['priority']</warning>;
}
function laterMode(bool $both) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    $first = <warning descr="Match the priority queue extraction mode.">$q->top()['priority']</warning>;
    if ($both) { $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH); }
    return $first;
}
function otherQueue(bool $both) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    $r = new SplPriorityQueue();
    if ($both) { $r->setExtractFlags(SplPriorityQueue::EXTR_BOTH); $r->{'insert'}('x', 1); }
    return <warning descr="Match the priority queue extraction mode.">$q->extract()['priority']</warning>;
}
function unprovenInsertOnly(bool $any) {
    $q = new SplPriorityQueue();
    if ($any) { $q->insert('task', 5); }
    return $q->extract()['priority'];
}
function conditionalArrayPriority(bool $rich) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    if ($rich) { $q->insert('other', ['priority' => 1]); }
    $q->setExtractFlags(SplPriorityQueue::EXTR_PRIORITY);
    return $q->extract()['priority'];
}
function separateScope() {
    $g = new SplPriorityQueue(); $g->setExtractFlags(SplPriorityQueue::EXTR_BOTH);
}
$g = new SplPriorityQueue(); $g->insert('task', 5);
echo <warning descr="Match the priority queue extraction mode.">$g->top()['priority']</warning>;
