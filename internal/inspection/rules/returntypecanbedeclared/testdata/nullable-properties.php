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
    public function <weak_warning descr="Declare ': ?int' as the return type.">getNumber</weak_warning>()
    {
        return $this->number;
    }

    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type.">getLabel</weak_warning>()
    {
        return $this->label;
    }

    /** @return Order */
    public function <weak_warning descr="Declare ': ?\Order' as the return type.">getOrder</weak_warning>()
    {
        return $this->order;
    }

    /** @return array */
    public function <weak_warning descr="Declare ': array' as the return type.">getLines</weak_warning>()
    {
        return $this->lines;
    }

    /** @return Order */
    public function <weak_warning descr="Declare ': ?\Order' as the return type.">getParent</weak_warning>()
    {
        return $this->parent;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">getYear</weak_warning>()
    {
        return $this->year;
    }

    /** @return Order */
    public function <weak_warning descr="Declare ': ?\Order' as the return type.">getOwner</weak_warning>()
    {
        return $this->owner;
    }
}
