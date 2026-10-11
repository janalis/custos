<?php
function keptInsideTry(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    try {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    } catch (Exception $e) {
    }
    foreach ($l as $item) {
    }
    return count($l);
}

function keptOnRemainingBranch(bool $stop): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    if ($stop) {
        return -1;
    } else {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    }
    foreach ($l as $item) {
    }
    return count($l);
}

function maybeKept(bool $keep): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    if ($keep) {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    }
    foreach ($l as $item) {
    }
    return count($l);
}

function keptByShortCircuit(bool $keep): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $keep && $l->setIteratorMode(0);
    foreach ($l as $item) {
    }
    return count($l);
}

function keptInLoopHeader(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    for ($i = 0; $i < 2; $l->setIteratorMode(0), $i++) {
    }
    foreach ($l as $item) {
    }
    return count($l);
}

function keptByClosure(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $keep = function () use ($l) {
        $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    };
    $keep();
    foreach ($l as $item) {
    }
    return count($l);
}

function keptByEarlierClosure(): int
{
    $l = new SplQueue();
    $keep = function (SplQueue $q) {
        $q->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
    };
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $keep($l);
    foreach ($l as $item) {
    }
    return count($l);
}

function capturedByArrowFunction(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $keep = fn () => resetMode($l);
    $keep();
    foreach ($l as $item) {
    }
    return count($l);
}

function keptThroughCallableArray(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $set = [$l, 'setIteratorMode'];
    $set(0);
    foreach ($l as $item) {
    }
    return count($l);
}

function keptThroughArrayElement(): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $all = ['list' => $l];
    $all['list']->setIteratorMode(0);
    foreach ($l as $item) {
    }
    return count($l);
}

function keptThroughDynamicName(string $method): int
{
    $l = new SplQueue();
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
    $other = new SplQueue();
    $other->$method(0);
    $l->$method(0);
    foreach ($l as $item) {
    }
    return count($l);
}

$l = new SplStack();
$l->setIteratorMode(SplDoublyLinkedList::IT_MODE_DELETE);
if (PHP_VERSION_ID > 0) {
    $l->setIteratorMode(SplDoublyLinkedList::IT_MODE_KEEP);
}
foreach ($l as $item) {
}
