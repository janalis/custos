<?php
namespace Shop;

class Tag {}

class BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes = <weak_warning descr="Explicit null default is redundant; remove it.">NULL</weak_warning>;
}

class Cart extends BaseCart
{
    protected $items = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">[]</weak_warning>;
    private $secret = 'x';
    public static $currency = 'EUR'; // own static storage: kept
    protected $tagClass = <weak_warning descr="Default repeats the inherited value; drop the re-declaration.">Tag::class</weak_warning>;
    protected $notes = 'n/a';
    private $mode // legacy
        = <weak_warning descr="Explicit null default is redundant; remove it.">\null</weak_warning>;

    private $total = <weak_warning descr="Default is always replaced by the constructor; remove it.">0</weak_warning>;
    private $lines = [];
    private $count = 0;
    private $owner;
    private $history = [];
    private $later = 'a';
    protected $shown = 1;

    public function __construct($owner)
    {
        $this->total = 100;
        $this->total = 200;
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->lines = [];</weak_warning>
        <weak_warning descr="Assignment writes the property's default value; remove it.">$this->owner = null;</weak_warning>
        $this->count = $this->count + 1;
        $this->history = array_merge($this->history, [$owner]);
        if ($owner) {
            $this->later = 'b';
        }
        $this->shown = 2;
    }

    public function reset()
    {
        $this->lines = [];
    }
}
