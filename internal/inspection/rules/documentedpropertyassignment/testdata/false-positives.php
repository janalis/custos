<?php
declare(strict_types=1);

namespace Shop;

class Order {}

final class Cart
{
    /** @var positive-int */
    public int $count = 1;
    /** @var Order */
    public int $stale = 0;
    public int $total = 0;
    /** @var int */
    public $loose;

    public function fill(int $n)
    {
        // Ranges and lengths count only against exact values.
        $this->count = $n;
        // A PHPDoc type the native type rejects is stale.
        $this->stale = 3;
        // Values the native types already reject are
        // InvalidPropertyAssignment's.
        $this->total = 'x';
        // Unknown values.
        $this->loose = $this->unknown();
    }
}

// Template-typed properties.
/** @template T */
final class Box
{
    /** @var T */
    public $value;

    public function put()
    {
        $this->value = 'x';
    }
}

// A receiver typed only by PHPDoc.
/** @param Cart $cart */
function documentedReceiver($cart) { $cart->loose = 'x'; }

// PHP's own classes: their PHPDoc is not a contract.
class Failure extends \Exception
{
    public function rename() { $this->message = 1; }
}

// A subtype that redeclares the property with another PHPDoc type.
class Ledger
{
    /** @var int */
    public $entry;
}

class Journal extends Ledger
{
    /** @var int|string */
    public $entry;
}

function post(Ledger $l) { $l->entry = 'x'; }
