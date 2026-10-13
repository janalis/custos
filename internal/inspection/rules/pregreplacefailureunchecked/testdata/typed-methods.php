<?php

function plain_text(string $text): void
{
}

final class Label
{
    public function __construct(private string $text = '')
    {
    }

    public static function of(string $text): self
    {
        return new self($text);
    }

    public static function maybe(?string $text): self
    {
        return new self($text ?? '');
    }

    public static function any($text): void
    {
    }

    public static function second($first, string $text): void
    {
    }

    public function rename(string $text): void
    {
        $this->text = $text;
    }
}

function staticMethodParameter($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);

    return <warning descr="Handle regex replacement failure before using its result.">Label::of($clean)</warning>;
}

function methodParameter($input, Label $label)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    <warning descr="Handle regex replacement failure before using its result.">$label->rename($clean)</warning>;
}

function nullsafeMethodParameter($input, ?Label $label)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    <warning descr="Handle regex replacement failure before using its result.">$label?->rename($clean)</warning>;
}

function namedStaticParameter($input)
{
    $clean = preg_replace_callback('/\s+/u', static fn (array $m): string => ' ', $input);
    <warning descr="Handle regex replacement failure before using its result.">Label::second(1, text: $clean)</warning>;
}

// The static call does not return for null: the later read is not reported again.
function reportedOnceAtTheCall($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    <warning descr="Handle regex replacement failure before using its result.">Label::of($clean)</warning>;

    return strlen($clean);
}

function guardedStaticParameter($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    if ($clean === null) {
        return null;
    }

    return Label::of($clean);
}

function nullableParameter($input)
{
    return Label::maybe(preg_replace('/\s+/u', ' ', $input));
}

function untypedParameter($input)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    Label::any($clean);
    Label::second($clean, 'x');
}

function unknownReceiver($input, $label)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    $label->rename($clean);
}

function computedMethod($input, Label $label, $name)
{
    $clean = preg_replace('/\s+/u', ' ', $input);
    $label->$name($clean);
    Label::$name($clean);
}
