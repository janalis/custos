<?php
namespace Gen;

abstract class Base
{
    protected $indentation = '    ';
    protected $glue = ', ';
    protected $sep = "a b";
    protected $quote = 'it\'s  x';
    protected $text = <<<'T'
        a  b
        T;
    /** @var list<string> */
    protected array $parts = [];
    protected array $items = [];
    protected $count = [1, 2];
}
