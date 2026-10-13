<?php
class Filler { public $q; function fill() { $this->q->enqueue('job'); } }
function depth(array $tokens) {
    $stack = new SplStack();
    foreach ($tokens as $t) {
        if ($t === ')') {
            $stack->pop();
        } else {
            $stack->push($t);
        }
    }
    return count($stack);
}
function later() {
    $h = new SplMinHeap();
    for ($i = 0; $i < 3; $i++) {
        if ($i > 0) { echo $h->extract(); }
        $h->insert($i);
    }
}
function generator() {
    $q = new SplQueue();
    while (true) {
        $v = yield;
        if ($v === null) { yield $q->dequeue(); } else { $q->enqueue($v); }
    }
}
function jump() {
    $q = new SplQueue();
    $n = 0;
    retry:
    if ($n++ > 0) { $q->pop(); }
    $q->push(1);
    if ($n < 2) goto retry;
}
function ternary() {
    $q = new SplQueue();
    return $q->isEmpty() ? null : $q->dequeue();
}
function shortCircuit() {
    $q = new SplQueue();
    $q->isEmpty() || $q->dequeue();
}
function drain() {
    $q = new SplQueue();
    while (!$q->isEmpty()) {
        $q->dequeue();
    }
}
function andGuard() {
    $q = new SplQueue();
    if (!$q->isEmpty() && $q->top() === 1) { echo "one"; }
    return $q->count() > 0 ? $q->bottom() : null;
}
function viaProperty(Filler $o) {
    $q = new SplQueue();
    $o->q = $q;
    $o->fill();
    return $q->dequeue();
}
function viaArray() {
    $q = new SplQueue();
    $holder = ['q' => $q];
    $holder['q']->enqueue('job2');
    return $q->dequeue();
}
function viaAlias() {
    $a = new SplQueue();
    $b = $a;
    $b->enqueue(1);
    return $a->dequeue();
}
