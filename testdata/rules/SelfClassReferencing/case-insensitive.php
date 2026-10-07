<?php
namespace Depot;

class Crate
{
    public function load(<weak_warning descr="Refer to the class as 'self' instead of 'Crate'.">crate</weak_warning> $other)
    {
        $a = new <weak_warning descr="Refer to the class as 'self' instead of 'Crate'.">CRATE</weak_warning>();
        $b = <weak_warning descr="Refer to the class as 'self' instead of 'Crate'.">\depot\crate</weak_warning>::LIMIT;
        $c = <weak_warning descr="Refer to the class as '__CLASS__' instead of 'Crate::class'.">Crate::CLASS</weak_warning>;
        $d = crate::class;
        return $a;
    }
}
