<?php
trait Greets
{
    public function hello() {}
}

trait Waves
{
    public function hello() {}
}

class Person
{
    public function greet() {}
    public function wave() {}
    public function shout() {}
}

class Speaker extends Person
{
    use Greets, Waves {
        Greets::hello insteadof Waves;
        hello as greet;
    }

    public function talk($name)
    {
        parent::greet();
        parent::$name();
        self::hello();
        <weak_warning descr="Call it as '$this->shout()' instead of through 'parent::'.">parent::shout()</weak_warning>;
    }
}
