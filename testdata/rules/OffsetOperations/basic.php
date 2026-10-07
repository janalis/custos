<?php
class Registry {
    /** @param string $name */
    public function offsetGet($name) { return null; }
}
class Dyn {
    public function __set($k, $v) {}
}

function demo(array $rows, int $limit) {
    $title = 'abc';
    echo $title[1];
    $title[<error descr="Index of type array does not fit the accepted string|int.">array_keys($rows)</error>] = 'z';
    $title[] = 'x';

    $count = 10;
    <error descr="'$count' does not support offset access (types: int).">$count[2]</error> = 1;
    <error descr="'$limit' does not support offset access (types: int).">$limit[0]</error>;

    $reg = new Registry();
    echo $reg['db'];
    echo $reg[<error descr="Index of type int does not fit the accepted string.">7</error>];

    $dyn = new Dyn();
    $dyn[<error descr="Index of type \ArrayObject does not fit the accepted string|int.">new ArrayObject()</error>] = 1;

    $plain = new DateTime();
    <error descr="'$plain' does not support offset access (types: \DateTime).">$plain[]</error> = 1;
}
