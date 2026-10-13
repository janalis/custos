<?php

final class Page
{
    public static function parse(string $markup): self
    {
        return new self();
    }

    public static function maybe(?string $markup): self
    {
        return new self();
    }

    public static function any($markup): void
    {
    }

    public function append(string $markup): void
    {
    }
}

function staticMethodParameter($path)
{
    $markup = file_get_contents($path);

    return <warning descr="Handle file reading failure before using its result.">Page::parse($markup)</warning>;
}

function methodParameter($stream, Page $page)
{
    $markup = stream_get_contents($stream);
    <warning descr="Handle file reading failure before using its result.">$page->append($markup)</warning>;
}

function namedMethodParameter($path, Page $page)
{
    <warning descr="Handle file reading failure before using its result.">$page->append(markup: file_get_contents($path))</warning>;
}

function guardedStaticParameter($path)
{
    $markup = file_get_contents($path);
    if ($markup === false) {
        return null;
    }

    return Page::parse($markup);
}

function nullableParameter($path)
{
    return Page::maybe(file_get_contents($path));
}

function untypedParameter($path)
{
    Page::any(file_get_contents($path));
}

function unknownReceiver($path, $page)
{
    $page->append(file_get_contents($path));
}

function computedMethod($path, Page $page, $name)
{
    $page->$name(file_get_contents($path));
    Page::$name(file_get_contents($path));
}
