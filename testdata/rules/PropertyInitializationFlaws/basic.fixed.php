<?php
namespace Shop;

class Tag {}

class BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes;
}

class Cart extends BaseCart
{
    protected $items = [];
    private $secret = 'x';
    public static $currency = 'EUR';
    protected $tagClass = Tag::class;
    protected $notes = 'n/a';
    private $mode;

    private $total;
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
        $this->lines = [];
        $this->owner = null;
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
