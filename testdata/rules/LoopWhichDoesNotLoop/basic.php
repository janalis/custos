<?php
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($queue as $job) {
    process($job);
    break;
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">for</warning> ($t = 0; $t < 5; $t++) {
    return $t;
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">while</warning> (poll()) {
    throw new \LogicException('stop');
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">do</warning> {
    break 1;
} while ($again);
while (wait()) {
    // the condition does the work: still loops
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($queue as $unused) {
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($grid as $line) {
    foreach ($line as $cell) {
        while (true) {
            continue 2;        // continues the middle loop, not the outer one
        }
    }
    break;
}

foreach ($grid as $line) {
    foreach ($line as $cell) {
        continue 2;            // continues the outer loop
    }
    break;
}
foreach ($grid as $line) {
    if (!$line) { continue; }
    break;
}
while ($ok) {
    step();
}
/** @var \Iterator $cursor */
foreach ($cursor as $first) {
    break;
}
while ($ok) break;

class Plain {}
function firstOfPlain(Plain $p, Bag2 $b) {
    <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($p as $v) {
        return $v;
    }
    <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($b as $v) {
        throw new \LogicException('x');
    }
}
class Bag2 implements \IteratorAggregate {
    public function getIterator(): \Iterator { return new \ArrayIterator([]); }
}
<warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($items as $it) {
    $f = function () { return 1; };
    break 1;
}
