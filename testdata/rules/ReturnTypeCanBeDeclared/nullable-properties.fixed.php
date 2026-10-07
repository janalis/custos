<?php

class Order {}

final class Invoice
{
    /** @var int */
    private $number;

    /** @var string */
    private $label = 'draft';

    /** @var Order */
    private $order = null;

    /** @var array */
    private $lines;

    /** @var Order */
    private $parent;

    private $owner;

    public function __construct(private int $year)
    {
        $this->lines = [];
        $this->parent = null;
    }

    /** @return int */
    public function getNumber(): ?int
    {
        return $this->number;
    }

    /** @return string */
    public function getLabel(): string
    {
        return $this->label;
    }

    /** @return Order */
    public function getOrder(): ?\Order
    {
        return $this->order;
    }

    /** @return array */
    public function getLines(): array
    {
        return $this->lines;
    }

    /** @return Order */
    public function getParent(): ?\Order
    {
        return $this->parent;
    }

    public function getYear(): int
    {
        return $this->year;
    }

    /** @return Order */
    public function getOwner(): ?\Order
    {
        return $this->owner;
    }
}
