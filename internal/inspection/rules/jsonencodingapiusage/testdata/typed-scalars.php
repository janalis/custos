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
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($text)</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($ratio)</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($mixed)</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($untyped)</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($ref)</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($written)</weak_warning>;
}

function calls(Signal $signal, ?Signal $maybe, \Countable $countable, $any, string $method): void
{
    echo json_encode([$signal->code(), $maybe?->code(), Signal::number(), exitCode()]);
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($signal->name())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($signal->missing())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($signal->$method())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($countable->count())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($any->code())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(label())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($method())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(Unknown::number())</weak_warning>;
    echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($any->code)</weak_warning>;
}

echo <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($top)</weak_warning>;
