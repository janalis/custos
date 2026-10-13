<?php
// Only the tested value of an is_*() call counts, and an object may be
// invokable.
namespace Review\IsChecks;

interface Job
{
    public function handle(): string;
}

class Open
{
}

final class Sealed
{
    public function same(self $other): bool
    {
        return <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_callable($other)</warning>;
    }
}

final class Invokable
{
    public function __invoke(): string
    {
        return 'invoke';
    }
}

enum Level
{
    case Low;
}

function run(Job $job, Open $open, Invokable $inv, Sealed $sealed, Level $level): array
{
    return [
        is_callable($job),
        is_callable($open),
        is_callable($inv),
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_callable($sealed)</warning>,
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_callable($level)</warning>,
    ];
}

// `self` outside a class names nothing: unknown.
function outside(self $x): bool
{
    return is_callable($x);
}

function flags($value, string $class, bool $allowString, bool $syntaxOnly, int $n): array
{
    return [
        is_a($value, $class, $allowString),
        is_callable($value, $syntaxOnly),
        is_callable($value, syntax_only: $syntaxOnly),
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_string(value: $n)</warning>,
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_a(object_or_class: $n, class: $class)</warning>,
        <warning descr="This check is always false for the declared parameter type; is the parameter being reused?">is_a($n, $class)</warning>,
    ];
}
