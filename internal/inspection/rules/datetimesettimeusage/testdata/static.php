<?php
class Stamp extends DateTime
{
    public function reset($usec)
    {
        parent::setTime(0, 0, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">$usec</error>);
        self::setTime(0, 0, 0, <error descr="Microseconds argument requires PHP 7.1+; on this version the call returns false.">$usec</error>);
        parent::setTime(0, 0, 0);
        return parent::modify('+1 day');
    }
}

class Loose
{
    public function reset()
    {
        return parent::setTime(0, 0, 0, 5);
    }
}
