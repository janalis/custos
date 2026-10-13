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
        self::honk();
        self::build();
    }

    public function drive($speed)
    {
        $this->honk();
        Car::start();
        $this->honk($speed, 2);
        $this->honk();
        parent::start();
        $this->honk();
        car::start();
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
$car->honk();
$car::build();
