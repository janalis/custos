<?php
namespace App;

// The parent is declared in another file: defaults are compared as text and
// a default referring to a class is never reported.
class Local extends \Lib\Remote
{
    protected $limit = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">10</weak_warning>;
    protected $mode = self::MODE;
}

class Base
{
    const KIND = 'base';
    protected $kind = self::KIND;
    protected $eol = [PHP_EOL];
    protected $up = 1;
}

class Derived extends Base
{
    const KIND = 'derived';
    // self:: names a different class here
    protected $kind = self::KIND;
    protected $eol = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">[PHP_EOL]</weak_warning>;
    protected $up = 1 + 0;
}

class NoPrivate
{
    public $a = 1;
    public function __construct() { $this->a = 2; }
}

class Ctor
{
    private int|string $id;
    private $plain;
    private $list = [];
    private $name = 'x';

    public function __construct($other, $n)
    {
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->id = null;</weak_warning>
        $this->plain = 1;
        $this->name .= 'y';
        $other->list = [1];
        $this->$n = [2];
        if ($n) { return; } else { echo 1; }
        $this->list = [3];
    }
}

class Grand { const G = 1; }
class Mid extends Grand { protected $g = parent::G; }
class Leaf extends Mid
{
    // parent:: names Mid here, Grand in Mid
    protected $g = parent::G;
}

// Not a constant expression (rejected by PHP, accepted by the parser): the
// dynamic class part is not a class reference.
class DynBase { protected $d = $x::C; }
class DynChild extends DynBase { protected $d = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">$x::C</weak_warning>; }

// Error recovery: a property with a missing default.
class Broken
{
    public $x = ;
}
