<?php
namespace App;

class BootTest
{
    public function a() { return realpath(__DIR__ . '/..'); }
}

class Boot
{
    public function a() { return <warning descr="Use 'dirname(__DIR__) . '/'' instead: realpath() fails inside stream wrappers.">realpath(__DIR__ . '/../')</warning>; }
}
