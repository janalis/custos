<?php

namespace Shop\Billing;

use Other\Tax;

// Tax\Rate would resolve to Other\Tax\Rate
$rate = \Shop\Billing\Tax\Rate::class;

namespace Shop\Billing\Tax;
class Rate {}

namespace Other;
class Tax {}
