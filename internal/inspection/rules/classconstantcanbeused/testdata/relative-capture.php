<?php

namespace Shop\Billing;

use Other\Tax;

// Tax\Rate would resolve to Other\Tax\Rate
$rate = <weak_warning descr="Use \Shop\Billing\Tax\Rate::class instead of the class name string.">'Shop\Billing\Tax\Rate'</weak_warning>;

namespace Shop\Billing\Tax;
class Rate {}

namespace Other;
class Tax {}
