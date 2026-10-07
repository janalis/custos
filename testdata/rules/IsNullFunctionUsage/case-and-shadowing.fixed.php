<?php
namespace Inventory;

function is_null($v) { return $v === 0; }

$own   = is_null($stock);
$own2  = !is_null($stock);
$built = $stock === null;

namespace Billing;

use function Inventory\is_null as zeroCheck;

$a = $total !== null;
$b = Other\is_null($total);
