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
    $win = fopen('NUL', 'c');
    $pipe = fopen('\\\\.\\pipe\\' . $path, 'r');
    $nul2 = fopen('nul.txt', 'rb');
    $devlike = fopen('/devices/' . $path, 'rb');
    $file = fopen('file://' . $path, 'rb');
    $rel = fopen('data/' . $path, 'rb');
    $local = fopen('notes.txt', 'rb');
    $num = fopen(0, 'rb');
    $odd = fopen('://x', 'rb');
    $wrapped = fopen(strtr(__CLASS__, '\\', '-') . '://' . $path, 'r');
    $scheme = fopen(($path) . ('://x'), 'r');
    $dir = fopen($path . '/x', 'rb');
    $twice = fopen($path . $fd . '://', 'rb');
    $named = fopen($path, 'rb');
    $nowdoc = fopen(<<<'P'
        php://temp
        P, 'w');
    $bad = fopen('php://memory', 'rb');
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

function misplaced_default(string $filename, string $mode = 'rb')
{
    return fopen($filename, $mode);
}
