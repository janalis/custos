<?php
class Courier
{
    public function send($to) { return true; }
}

class LocalCourier extends Courier
{
    }
