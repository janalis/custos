<?php
interface Receiver { public function get(int $n): iterable; }

function receive(Receiver $r, iterable $items, ?iterable $maybe, iterable|string $either, \Traversable $t) {
    [$first, $second] = $r->get(2);      // custos: an iterable may be an array
    [$a] = $items;
    [$b] = $maybe;
    <error descr="Destructuring a value that is not an array.">[$c] = $either</error>;
    <error descr="Destructuring a value that is not an array.">[$d] = $t</error>;
    return [$first, $second, $a, $b, $c, $d];
}
