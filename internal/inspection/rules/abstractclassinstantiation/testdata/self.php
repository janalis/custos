<?php
abstract class Shape
{
    public static function make()
    {
        return <error descr="Instantiate a concrete class.">new self()</error>;
    }

    public function copy()
    {
        if ($this->ready()) {
            return [<error descr="Instantiate a concrete class.">new SELF</error>];
        }
        return null;
    }

    abstract protected function ready();
}
