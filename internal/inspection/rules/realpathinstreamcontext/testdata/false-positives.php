<?php
namespace App\Tests\Unit;

class Loader
{
    public function load()
    {
        return realpath(__DIR__ . '/../fixtures');
    }
}
