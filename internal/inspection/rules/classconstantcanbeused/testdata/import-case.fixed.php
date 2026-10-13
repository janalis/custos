<?php

namespace App;

use Other\LEDGER;

class Desk {
    public function names() {
        // `Ledger` is already taken by the import (alias names are
        // case-insensitive): no second import, stays fully qualified.
        return [\Vendor\Kit\Ledger::class];
    }
}

namespace Vendor\Kit;
class Ledger {}

namespace Other;
class Ledger {}
