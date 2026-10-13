<?php

class Engine
{
    public function start() {}
}

class Car extends Engine
{
    public function start() {}
    public function honk() {}
    public static function build() {}

    public static function factory()
    {
        <warning descr="Call 'honk' on an instance with '->' instead of '::'.">self::honk()</warning>;
        self::build();
    }

    public function drive($speed)
    {
        <warning descr="Call 'honk' with '$this->' instead of '::'.">static::honk()</warning>;
        <warning descr="Call 'start' with '$this->' instead of '::'.">Car::start()</warning>;
        <warning descr="Call 'honk' with '$this->' instead of '::'.">self::honk($speed, 2)</warning>;
        <warning descr="Call 'honk' on an instance with '->' instead of '::'.">$this::honk()</warning>;
        parent::start();
        <warning descr="Call 'honk' with '$this->' instead of '::'.">SELF::honk()</warning>;
        <warning descr="Call 'start' with '$this->' instead of '::'.">car::start()</warning>;
        $cb = function () { self::honk(); };
    }
}

class SportsCar extends Car
{
    public function start()
    {
        Car::start();
        Engine::start();
    }
}

$car = new Car();
<warning descr="Call 'honk' on an instance with '->' instead of '::'.">$car::honk()</warning>;
$car::build();
