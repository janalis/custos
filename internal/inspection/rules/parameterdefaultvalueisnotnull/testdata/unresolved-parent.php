<?php

class FiltersTestBase extends \Drupal\KernelTests\KernelTestBase
{
    protected function setUp($importViews = TRUE): void {}
}

class Known {}

class Middle extends Known
{
    use \Missing\Helpers;
}

class Leaf extends Middle
{
    public function render($mode = 'full') {}
}

class Plain extends Known
{
    public function render(<weak_warning descr="Prefer null as the default value for this parameter.">$mode = 'full'</weak_warning>) {}
}

class HelperTest extends \PHPUnit\Framework\TestCase
{
    public function __construct(<weak_warning descr="Prefer null as the default value for this parameter.">$name = ''</weak_warning>) {}

    private function builder(<weak_warning descr="Prefer null as the default value for this parameter.">$multiple = false</weak_warning>) {}
}

class LeafHelper extends Middle
{
    private function builder(<weak_warning descr="Prefer null as the default value for this parameter.">$multiple = false</weak_warning>) {}
}
