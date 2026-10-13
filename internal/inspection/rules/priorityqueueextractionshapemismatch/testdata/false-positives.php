<?php
function conditionalMode(bool $both) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    if ($both) { $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH); }
    return $both ? $q->extract()['priority'] : $q->extract();
}
function conditionalArrayInsert(bool $rich) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    if ($rich) { $q->insert(['priority' => 9], 6); }
    return $q->extract()['priority'];
}
function arrowMode() {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    $both = fn() => $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH);
    $both();
    return $q->extract()['priority'];
}
function closureMode() {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    $both = function () use ($q) { $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH); };
    $both();
    return $q->extract()['priority'];
}
function loopMode(array $rounds) {
    $q = new SplPriorityQueue(); $q->insert('a', 1); $q->insert('b', 2);
    foreach ($rounds as $round) {
        if ($round) { echo $q->top()['priority']; }
        $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH);
    }
}
function shortCircuitMode(bool $both) {
    $q = new SplPriorityQueue(); $q->insert('task', 5);
    $both && $q->setExtractFlags(SplPriorityQueue::EXTR_BOTH);
    return $both ? $q->top()['data'] : null;
}
$g = new SplPriorityQueue(); $g->insert('task', 5);
if ($argc > 1) { $g->setExtractFlags(SplPriorityQueue::EXTR_BOTH); }
echo $argc > 1 ? $g->top()['priority'] : '';
