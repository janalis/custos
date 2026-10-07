<?php

namespace App;

use Other\LEDGER;

class Desk {
    public function names() {
        // `Ledger` is already taken by the import (alias names are
        // case-insensitive): no second import, stays fully qualified.
        return [<weak_warning descr="Use \Vendor\Kit\Ledger::class instead of the class name string.">'Vendor\Kit\Ledger'</weak_warning>];
    }
}

namespace Vendor\Kit;
class Ledger {}

namespace Other;
class Ledger {}
