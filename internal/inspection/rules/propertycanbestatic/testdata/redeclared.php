<?php
class Generator
{
    protected $decoded = ['%2F' => '/', '%40' => '@', '%3A' => ':'];
    private $kept = ['a', 'b', 'c'];
    protected <weak_warning descr="Large array default on an instance property; consider a static property or a class constant.">$plain</weak_warning> = ['x', 'y', 'z'];
}

class Custom extends Generator
{
    protected $decoded = ['%2F' => '/'];
}

class Deeper extends Custom {}

final class Leaf extends Deeper
{
    private $kept = [];
}
