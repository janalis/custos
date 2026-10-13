<?php
namespace Billing;

class Invoice {}
class Ledger {}

class Narrow {
    public function a(Invoice $i): invoice { return new INVOICE(); }
    public function b(\Billing\ledger $l): Ledger { return $l; }
}

class <weak_warning descr="Depends on 3 distinct classes; consider splitting it up.">Wide</weak_warning> {
    public function a(Invoice $i): invoice { return new \ArrayObject(); }
    public function b(LEDGER $l): Ledger { return $l; }
}
