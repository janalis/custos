<?php

namespace Vendor\Kit {
    class Meter {}
    class Gauge {}
    class Valve {}
    class Pump {}
    class Probe {}
    class Sub {}
    interface Dial {}
    function Probe() {}
}

namespace First {
    // the import is added to this block, not to the first namespace of the file
    echo <weak_warning descr="Use \Vendor\Kit\Gauge::class instead of the class name string.">'Vendor\Kit\Gauge'</weak_warning>;
}

namespace Second {
    use Vendor\Kit\Meter as M;
    use vendor\kit\valve;
    use function Vendor\Kit\Probe;

    echo <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'Vendor\Kit\Meter'</weak_warning>;
    // valve::class would return the spelling of the import
    echo <weak_warning descr="Use \Vendor\Kit\Valve::class instead of the class name string.">'Vendor\Kit\Valve'</weak_warning>;
    // a function import is not a class import
    echo <weak_warning descr="Use \Vendor\Kit\Probe::class instead of the class name string.">'Vendor\Kit\Probe'</weak_warning>;
}

namespace Third {
    class M {}
    interface Dial {}

    // the alias M of the previous block means Third\M here
    echo <weak_warning descr="Use \Vendor\Kit\Meter::class instead of the class name string.">'Vendor\Kit\Meter'</weak_warning>;
    // Dial is declared in this block, Pump is referenced by it
    echo <weak_warning descr="Use \Vendor\Kit\Dial::class instead of the class name string.">'Vendor\Kit\Dial'</weak_warning>;
    echo <weak_warning descr="Use \Vendor\Kit\Pump::class instead of the class name string.">'Vendor\Kit\Pump'</weak_warning>;
    echo Pump::START;
}

namespace Fourth {
    use Other\Item;

    class Item2 {}

    class Holder {
        public function names() {
            return [
                // Item alone is Other\Item in this block
                <weak_warning descr="Use Item::class instead of the class name string.">__NAMESPACE__ . '\Item'</weak_warning>,
                // 'FourthItem2' at runtime
                __NAMESPACE__ . 'Item2',
                <weak_warning descr="Use Sub\Leaf::class instead of the class name string.">__NAMESPACE__ . '\Sub\Leaf'</weak_warning>,
                // an import of Sub would capture the Sub\Leaf written above
                <weak_warning descr="Use \Vendor\Kit\Sub::class instead of the class name string.">'Vendor\Kit\Sub'</weak_warning>,
            ];
        }
    }
}

namespace Fourth\Sub {
    class Leaf {}
}

namespace Fourth {
    class Item {}
}

namespace Other {
    class Item {}
}
