<?php
namespace App\Tests\Fixtures;

class ConstructorDummy
{
    private $timezone;

    public function __construct(\DateTimeZone $timezone)
    {
        $this->timezone = $timezone;
    }
}
