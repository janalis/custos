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
