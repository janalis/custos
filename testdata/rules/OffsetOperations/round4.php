<?php
interface Container {}
abstract class BaseBag {}
abstract class SealedBag { }
final class Plain {}
class Registry
{
    private iterable $items = [];

    public function add(string $k, $v): void
    {
        $this->items[$k] = $v;
    }
}

function viaInterface(Container $c, BaseBag $b, Plain $p)
{
    return [$c['config'], $b['x'], <error descr="'$p' does not support offset access (types: \Plain).">$p['y']</error>];
}

/** @return mixed|Plain|null */
function anyKey() {}
/** @return string|void */
function maybeKey() {}

function keys(array $a)
{
    return [isset($a[anyKey()]), $a[maybeKey()]];
}

function counts(): array|int { return 0; }

function quiet()
{
    $c = counts();
    $a = $c[1] ?? 0;
    $b = isset($c[2]['x']);
    $d = empty($c[3]);
    $e = isset($c[4]->p);
    $g = isset($a[<error descr="'$c' does not support offset access (types: array|int).">$c[7]</error>]);
    $h = isset($a->{<error descr="'$c' does not support offset access (types: array|int).">$c[8]</error>});
    $f = $a ?? <error descr="'$c' does not support offset access (types: array|int).">$c[5]</error>;
    return [$a, $b, $d, $e, $f, $g, $h, <error descr="'$c' does not support offset access (types: array|int).">$c[6]</error>];
}
