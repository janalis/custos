<?php
class Courier
{
    public function send($to) { return true; }
}

class LocalCourier extends Courier
{
    public function <weak_warning descr="Method 'send' only forwards to its parent; remove it.">send</weak_warning>($to)
    {
        return PARENT::Send($to);
    }
}
