<?php
// The binary flag only matters for plain files.
function streams(int $fd, $path)
{
    $out = fopen('php://stdout', 'w');
    $mem = fopen(('php://memory'), 'r+');
    $dup = fopen('php://fd/' . $fd . '/x', 'r+');
    $zval = fopen("Wrapper-Name://", 'r+');
    $null = fopen('/dev/null', 'c');
    $tty = fopen('/dev/pts/' . $fd, 'w');
    $devlike = fopen('/devices/' . $path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $file = fopen('file://' . $path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $rel = fopen('data/' . $path, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $local = fopen('notes.txt', <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $num = fopen(0, <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $odd = fopen('://x', <warning descr="Add the 'b' flag to the mode for binary-safe file access.">'r'</warning>);
    $nowdoc = fopen(<<<'P'
        php://temp
        P, 'w');
    $bad = fopen('php://memory', <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">'br'</error>);
}

// A parameter's default only says what callers get when they pass nothing.
function open_file(string $filename, string $mode = 'r')
{
    return fopen($filename, $mode);
}

final class Writer
{
    private string $mode;

    public function __construct(private string $url, string $mode = 'a')
    {
        $this->mode = $mode;
    }

    public function open()
    {
        return fopen($this->url, $this->mode);
    }
}

function misplaced_default(string $filename, string $mode = 'br')
{
    return fopen($filename, <error descr="Move the 'b' flag to the end of the mode (e.g. 'rb', 'rb+').">$mode</error>);
}
