<?php

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;
use Vendor\Kit\Meter;

class Invoice {}
class Ledger {}

namespace Shop\Billing\Tax;
class Rate {}

namespace Vendor\Kit;
class Ledger {}
class Meter {}

namespace Shop\Billing;

class Clerk {
    public function names() {
        return [
            static::class,
            get_parent_class(),
            Bag::class,
            Bag::class,
            Bill::class,
            \Shop\Billing\Ledger::class,
            Ledger::class,
            Tax\Rate::class,
            \Vendor\Kit\Ledger::class,
            Meter::class,
            Meter::class,
        ];
    }

    public function untouched($suffix) {
        $s  = '';
        $s .= '\ArrayObject';
        class_alias($suffix, '\Shop\Billing\Ledger');
        return [
            get_called_class($this),
            'arrayobject',
            'Bag',
            'Invoice',
            'shop\billing\invoice',
            '\ArrayObject' . $suffix,
            "\\ArrayObject{$suffix}",
            $suffix == 'Shop\Billing\Ledger',
            'Shop\Billing\Missing',
            'Not a class',
        ];
    }
}
