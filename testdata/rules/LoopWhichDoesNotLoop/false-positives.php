<?php
function gen() { yield 1; }
function first(\Generator $g) {
    foreach ($g as $v) {
        return $v;
    }
    return null;
}
foreach ($items as $it) {
    if ($it) {
        break;
    }
}
foreach ($items as $it) {
    exit(1);
}
foreach ($items as $it):
    work($it);
endforeach;
foreach ($items as $it) {
    $f = function () { foreach ([1] as $x) { continue; } };
    process($it);
}

class Bag implements \IteratorAggregate {
    public function getIterator(): \Iterator { return new \ArrayIterator([]); }
}
interface Cursor extends \Iterator {}
function firstOfBag(Bag $bag, Cursor $cursor, \ArrayIterator $it) {
    foreach ($bag as $v) {
        return $v;
    }
    foreach ($cursor as $v) {
        break;
    }
    foreach ($it as $v) {
        return $v;
    }
    return null;
}

class Ledger {
    public function entries() { yield ['id' => 1]; yield ['id' => 2]; }
    public function firstEntry() {
        foreach ($this->entries() as $entry) {
            return $entry;
        }
        return null;
    }
}
trait Paged {
    public function pages() { yield from [1, 2]; }
}
class Report {
    use Paged;
    public function firstPage() {
        foreach (($this->pages()) as $page) {
            break;
        }
    }
}
function firstOfLedger(Ledger $ledger) {
    foreach ($ledger->entries() as $entry) {
        return $entry;
    }
    return null;
}
