<?php
// XMLReader::open()/XML() are static in the stubs but open the instance
// they are called on.
class Dump
{
    private XMLReader $reader;

    public function __construct()
    {
        $this->reader = new XMLReader();
    }

    public function load(string $f, string $x): void
    {
        if (!$this->reader->open($f)) {
            return;
        }
        $this->reader->XML($x);
        $r = new XMLReader();
        $r->open($f);
    }
}
