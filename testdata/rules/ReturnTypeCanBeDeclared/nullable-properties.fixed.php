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
    public function getNumber()
    {
        return $this->number;
    }

    /** @return string */
    public function getLabel()
    {
        return $this->label;
    }

    /** @return Order */
    public function getOrder()
    {
        return $this->order;
    }

    /** @return array */
    public function getLines()
    {
        return $this->lines;
    }

    /** @return Order */
    public function getParent()
    {
        return $this->parent;
    }

    public function getYear(): int
    {
        return $this->year;
    }

    /** @return Order */
    public function getOwner()
    {
        return $this->owner;
    }
}
