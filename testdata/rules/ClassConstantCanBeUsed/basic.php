<?php

namespace Shop\Billing;

use Shop\Billing\Invoice as Bill;
use \ArrayObject as Bag;

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
            <weak_warning descr="Use static::class instead.">get_called_class()</weak_warning>,
            get_parent_class(),
            <weak_warning descr="Use \ArrayObject::class instead of the class name string.">'\ArrayObject'</weak_warning>,
            <weak_warning descr="Use \ArrayObject::class instead of the class name string.">'ArrayObject'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Invoice::class instead of the class name string.">'Shop\Billing\Invoice'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Ledger::class instead of the class name string.">"\\Shop\\Billing\\Ledger"</weak_warning>,
            <weak_warning descr="Use Ledger::class instead of the class name string.">__NAMESPACE__ . '\Ledger'</weak_warning>,
            <weak_warning descr="Use \Shop\Billing\Tax\Rate::class instead of the class name string.">'Shop\Billing\Tax\Rate'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Ledger::class instead of the class name string.">'Vendor\Kit\Ledger'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'Vendor\Kit\Meter'</weak_warning>,
            <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'\Vendor\Kit\Meter'</weak_warning>,
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
