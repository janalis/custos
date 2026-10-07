<?php
class Ledger
{
    public function open()
    {
        tag(__CLASS__);
        return new Ledger;
    }
}
