<?php
namespace App\Tests;

abstract class TestForProxyHelper
{
    public function &foo3(\stdClass &$b, $c) {}

    public function make()
    {
        $o = &new \stdClass();
    }
}
