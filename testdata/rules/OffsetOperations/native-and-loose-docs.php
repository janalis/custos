<?php
class NodeListChild extends DOMNodeList {}

function nodes(DOMDocument $d, ResourceBundle $b, DOMElement $e, NodeListChild $c)
{
    $l = $d->getElementsByTagName('a');
    echo $l[0], $b['calendar'], $e->attributes['id'], $c[1];
}

/**
 * @return array|int|string|float|bool
 */
function query($q)
{
    return $q;
}

function useQuery(): void
{
    $r = query('x');
    echo $r['data'];
}

function nativeScalar(int|array $v): void
{
    echo <error descr="'$v' does not support offset access (types: array|int).">$v['k']</error>;
}

/**
 * @param \stdClass $o
 */
function wrongDoc($o): void
{
    echo <error descr="'$o' does not support offset access (types: \stdClass).">$o['k']</error>;
}
