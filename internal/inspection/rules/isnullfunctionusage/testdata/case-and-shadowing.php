<?php
namespace Inventory;

function is_null($v) { return $v === 0; }

$own   = is_null($stock);
$own2  = !is_null($stock);
$built = <weak_warning descr="Replace with '$stock === null'.">\IS_NULL($stock)</weak_warning>;

namespace Billing;

use function Inventory\is_null as zeroCheck;

$a = <weak_warning descr="Replace with '$total !== null'.">!Is_Null($total)</weak_warning>;
$b = Other\is_null($total);
