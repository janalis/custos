<?php
namespace Depot;

class Crate
{
    public function load(self $other)
    {
        $a = new self();
        $b = self::LIMIT;
        $c = __CLASS__;
        $d = crate::class;
        return $a;
    }
}
