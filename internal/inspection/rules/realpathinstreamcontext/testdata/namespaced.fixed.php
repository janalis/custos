<?php
namespace App;

class BootTest
{
    public function a() { return realpath(__DIR__ . '/..'); }
}

class Boot
{
    public function a() { return dirname(__DIR__); }
}
