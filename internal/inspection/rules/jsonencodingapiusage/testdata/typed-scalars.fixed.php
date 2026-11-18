<?php
// Values declared as integers, booleans or null always encode.
final class Signal
{
    public function code(): int|false
    {
        return false;
    }

    public static function number(): ?int
    {
        return null;
    }

    public function name(): string
    {
        return 'TERM';
    }

    public function __call(string $name, array $args): int
    {
        return 0;
    }
}

function exitCode(): int
{
    return 0;
}

function label(): string
{
    return 'x';
}

function params(int $signal, int|false $previous = 0, ?bool $flag = null, true $on = true, string $text = '', float $ratio = 0.0, int|string $mixed = 0, $untyped = 0, int &$ref = 0, int $written = 0): void
{
    $written = label();
    echo json_encode(['exit', $signal, $previous, $flag, $on]);
    echo json_encode($text, JSON_THROW_ON_ERROR);
    echo json_encode($ratio, JSON_THROW_ON_ERROR);
    echo json_encode($mixed, JSON_THROW_ON_ERROR);
    echo json_encode($untyped, JSON_THROW_ON_ERROR);
    echo json_encode($ref, JSON_THROW_ON_ERROR);
    echo json_encode($written, JSON_THROW_ON_ERROR);
}

function calls(Signal $signal, ?Signal $maybe, \Countable $countable, $any, string $method): void
{
    echo json_encode([$signal->code(), $maybe?->code(), Signal::number(), exitCode()]);
    echo json_encode($signal->name(), JSON_THROW_ON_ERROR);
    echo json_encode($signal->missing(), JSON_THROW_ON_ERROR);
    echo json_encode($signal->$method(), JSON_THROW_ON_ERROR);
    echo json_encode($countable->count(), JSON_THROW_ON_ERROR);
    echo json_encode($any->code(), JSON_THROW_ON_ERROR);
    echo json_encode(label(), JSON_THROW_ON_ERROR);
    echo json_encode($method(), JSON_THROW_ON_ERROR);
    echo json_encode(Unknown::number(), JSON_THROW_ON_ERROR);
    echo json_encode($any->code, JSON_THROW_ON_ERROR);
}

echo json_encode($top, JSON_THROW_ON_ERROR);
