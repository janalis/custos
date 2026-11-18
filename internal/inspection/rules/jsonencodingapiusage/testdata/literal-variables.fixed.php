<?php
function literalVariables(string $file, string $raw, bool $cond): void
{
    // every value reaching the read is a literal that encodes
    $baseline = [['location' => 'A::b', 'count' => 1]];
    file_put_contents($file, json_encode($baseline));

    $grown = ['a' => 1];
    $grown[] = $raw;
    file_put_contents($file, json_encode($grown, JSON_THROW_ON_ERROR));

    $either = $cond ? ['a'] : $raw;
    file_put_contents($file, json_encode($either, JSON_THROW_ON_ERROR));

    file_put_contents($file, json_encode($raw, JSON_THROW_ON_ERROR));
}

function capturedVariables(string $raw): array
{
    // a by-value import holds the enclosing variable's literal
    $key = 'projects/demo';
    $plain = static function () use ($key): string {
        return json_encode(['name' => $key . '/v1']);
    };
    $param = static function (string $key): string {
        return json_encode($key, JSON_THROW_ON_ERROR);
    };
    $other = static function () use ($raw): string {
        return json_encode($raw, JSON_THROW_ON_ERROR);
    };
    $arrow = static fn (): string => json_encode($key, JSON_THROW_ON_ERROR);
    return [$plain, $param, $other, $arrow];
}

function referencedImport(): callable
{
    $key = 'projects/demo';
    return function () use (&$key): string {
        return json_encode(['name' => $key], JSON_THROW_ON_ERROR);
    };
}

function rewrittenImport(string $raw): callable
{
    $key = 'projects/demo';
    return static function () use ($key, $raw): string {
        $key .= $raw;
        return json_encode($key, JSON_THROW_ON_ERROR);
    };
}

class Conn
{
    public function add(string $body, array $headers): string { return $body; }
    public function fill(array &$headers): void { $headers[] = "\xB1"; }
}

function nullableReceiver(?Conn $c): array
{
    // a method called on null throws before binding its arguments
    $headers = ['type' => 'a'];
    $c->add('x', $headers);
    $filled = ['type' => 'a'];
    $c->fill($filled);
    return [json_encode(['headers' => $headers]), json_encode($filled, JSON_THROW_ON_ERROR)];
}

function integerVariables(string $raw, int $n): array
{
    // integer operations on locals holding integers; a value reaching
    // itself through a loop is not followed
    $time = time();
    $claims = ['iat' => $time + 3, 'exp' => -$time * 2];
    $step = 0;
    while ($n--) {
        $step = $step + 1;
    }
    $text = $raw;
    return [
        json_encode($claims),
        json_encode(['n' => $step - 1], JSON_THROW_ON_ERROR),
        json_encode(['n' => $text + 1], JSON_THROW_ON_ERROR),
        json_encode(['n' => $n * 2], JSON_THROW_ON_ERROR),
    ];
}
