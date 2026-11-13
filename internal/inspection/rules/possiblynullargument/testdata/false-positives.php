<?php
declare(strict_types=1);

namespace Shop;

class Order {}
class Coupon {}

function ship(Order $order) {}
function keep(?Order $order) {}
function maybe(Order $order = null) {}
function any(mixed $value) {}
function grow(?Order &$order) {}
function find(): ?Order { return null; }
function coupon(): ?Coupon { return null; }

class Desk
{
    private ?Order $current = null;

    public function refresh(): void {}

    public function run(?Order $order, ?Order $other, $unknown): void
    {
        // Parameters that accept null.
        keep($order);
        maybe($order);
        any($order);
        grow($order);
        // Rejected whatever the value: InvalidArgumentType's.
        ship(null);
        ship(coupon());
        // Unknown values.
        ship($unknown);
        // Checked.
        if ($other !== null) {
            ship($other);
        }
        if ($this->current === null) {
            return;
        }
        $this->refresh();
        ship($this->current);
        if (find() !== null) {
            ship(find());
        }
        // After an unpacked argument.
        ship(...[$order]);
    }

    /** @param Order|null $order */
    public function documented($order): void
    {
        ship($order);
    }

    public function bound(): \Closure
    {
        return function () {
            ship($this->current);
            $this->send($this->current);
        };
    }

    public function send(Order $order): void {}
}
