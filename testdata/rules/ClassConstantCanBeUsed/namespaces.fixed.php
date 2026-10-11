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
    use Vendor\Kit\Gauge;

echo Gauge::class;
}

namespace Second {
    use Vendor\Kit\Meter as M;
    use vendor\kit\valve;
    use function Vendor\Kit\Probe;
use Vendor\Kit\Probe;

    echo M::class;
    // valve::class would return the spelling of the import
    echo \Vendor\Kit\Valve::class;
    // a function import is not a class import
    echo Probe::class;
}

namespace Third {
    use Vendor\Kit\Meter;

class M {}
    interface Dial {}

    // the alias M of the previous block means Third\M here
    echo Meter::class;
    // Dial is declared in this block, Pump is referenced by it
    echo \Vendor\Kit\Dial::class;
    echo \Vendor\Kit\Pump::class;
    echo Pump::START;
}

namespace Fourth {
    use Other\Item;

    class Item2 {}

    class Holder {
        public function names() {
            return [
                // Item alone is Other\Item in this block
                namespace\Item::class,
                // 'FourthItem2' at runtime
                __NAMESPACE__ . 'Item2',
                Sub\Leaf::class,
                // an import of Sub would capture the Sub\Leaf written above
                \Vendor\Kit\Sub::class,
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
