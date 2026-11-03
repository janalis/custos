<?php
final class Page
{
    public function append(string $markup): void {}
}

// Reading back PHP's own memory stream cannot fail.
function memory(string $data, Page $page): void
{
    $buffer = fopen('php://memory', 'r+');
    fwrite($buffer, $data);
    rewind($buffer);
    $markup = stream_get_contents($buffer);
    $page->append($markup);
}

// Reading a file stream can.
function file(string $path, Page $page): void
{
    $handle = fopen($path, 'r');
    if (false === $handle) {
        return;
    }
    $markup = stream_get_contents($handle);
    <warning descr="Handle file reading failure before using its result.">$page->append($markup)</warning>;
}

// A private property holding only a memory stream reads like one.
final class Buffered
{
    private $stream;
    private $typed = null;
    public $open;
    private $file;
    private $moved;
    private $shared;
    private $cleared;
    private $none;
    private $passed;
    private $opaque;
    private $listed = [];
    private $spread;

    public function __construct(private Page $page, string $path)
    {
        $this->stream = fopen('php://memory', 'r+');
        $this->typed = fopen('php://temp', 'r+');
        $this->open = fopen('php://memory', 'r+');
        $this->file = fopen($path, 'r');
        $this->moved = fopen('php://memory', 'r+');
        $this->shared = fopen('php://memory', 'r+');
        $this->cleared = fopen('php://memory', 'r+');
        $this->passed = fopen('php://memory', 'r+');
        $this->opaque = fopen('php://memory', 'r+');
        $this->listed = fopen('php://memory', 'r+');
        $this->spread = fopen('php://memory', 'r+');
    }

    public function read(object $other, callable $sink): void
    {
        rewind($this->stream);
        $this->page->append(stream_get_contents($this->stream));
        $this->page->append(stream_get_contents($this->typed));
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->open))</warning>;
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->file))</warning>;
        $this->moved .= 'x';
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->moved))</warning>;
        $alias = &$this->shared;
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->shared))</warning>;
        unset($this->cleared);
        $this->page->append(stream_get_contents($this->cleared));
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->none))</warning>;
        settype($this->passed, 'string');
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->passed))</warning>;
        $sink($this->opaque);
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->opaque))</warning>;
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($other->stream))</warning>;
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this?->stream))</warning>;
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->listed))</warning>;
        [$this->spread] = [fopen('php://stdin', 'r')];
        <warning descr="Handle file reading failure before using its result.">$this->page->append(stream_get_contents($this->spread))</warning>;
        $copy = $this->stream;
    }
}

function outside(Page $page, object $o): void
{
    <warning descr="Handle file reading failure before using its result.">$page->append(stream_get_contents($o->stream))</warning>;
}

final class Undeclared
{
    public function read(Page $page): void
    {
        $this->stream = fopen('php://memory', 'r+');
        <warning descr="Handle file reading failure before using its result.">$page->append(stream_get_contents($this->stream))</warning>;
    }
}
