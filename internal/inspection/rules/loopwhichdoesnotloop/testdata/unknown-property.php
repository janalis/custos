<?php
class Invoice
{
    /** @ORM\ManyToMany(targetEntity="Credit", mappedBy="invoices") */
    protected $credits;
    protected array $lines = [];

    public function firstCredit()
    {
        foreach ($this->credits as $credit) {
            return $credit;
        }
        return null;
    }

    public function firstLine()
    {
        <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($this->lines as $line) {
            return $line;
        }
        return null;
    }
}
