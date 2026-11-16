<?php
namespace Shop;

function fill(array &$into): void {}

// Writing an offset of null makes it an array: no dereference. Reading
// one, or calling a method, still dereferences the value.
class Store
{
    public function write(?\ArrayObject $o, ?\ArrayObject $p, ?\ArrayObject $q, ?\ArrayObject $r, ?\ArrayObject $s): void
    {
        $o['k'] = 1;
        $o[] = 2;
        $o['a']['b'] = 3;
        $o['n'] .= 'x';
        $o['m'] ??= 4;
        $o['c']++;
        [$o['d'], $o['e']] = [5, 6];
        foreach ([7] as $o['f']) {
        }
        $ref = &$o['g'];
        unset($o['h']);
        fill($o['i']);

        echo <warning descr="Possible null dereference.">$p</warning>['k'];
        $p['k'] = 1;

        $x = <warning descr="Possible null dereference.">$q</warning>['k'] + 1;

        <warning descr="Possible null dereference.">$r</warning>->count();

        unknown_fn(<warning descr="Possible null dereference.">$s</warning>['k']);
    }
}
