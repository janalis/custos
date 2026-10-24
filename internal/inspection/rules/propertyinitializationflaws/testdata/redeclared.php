<?php
namespace Gen;

final class Child extends Base
{
    // Whitespace inside a string is part of the value.
    protected $indentation = '';
    protected $glue = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">', '</weak_warning>;
    protected $sep = "ab";
    protected $quote = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">'it\'s  x'</weak_warning>;
    protected $text = 'a  b';
    // A re-declaration narrowing the type is not a duplicate.
    /** @var list<int> */
    protected array $parts = [];
    /** @var list<int> */
    protected array $items = [];
    protected $count = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">[ 1,2 ]</weak_warning>;
}
