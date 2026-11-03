<?php
function literalVariables(string $file, string $raw, bool $cond): void
{
    // every value reaching the read is a literal that encodes
    $baseline = [['location' => 'A::b', 'count' => 1]];
    file_put_contents($file, json_encode($baseline));

    $grown = ['a' => 1];
    $grown[] = $raw;
    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($grown)</weak_warning>);

    $either = $cond ? ['a'] : $raw;
    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($either)</weak_warning>);

    file_put_contents($file, <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($raw)</weak_warning>);
}

function capturedVariables(string $raw): array
{
    // a by-value import holds the enclosing variable's literal
    $key = 'projects/demo';
    $plain = static function () use ($key): string {
        return json_encode(['name' => $key . '/v1']);
    };
    $param = static function (string $key): string {
        return <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($key)</weak_warning>;
    };
    $other = static function () use ($raw): string {
        return <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($raw)</weak_warning>;
    };
    $arrow = static fn (): string => <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($key)</weak_warning>;
    return [$plain, $param, $other, $arrow];
}

function referencedImport(): callable
{
    $key = 'projects/demo';
    return function () use (&$key): string {
        return <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode(['name' => $key])</weak_warning>;
    };
}

function rewrittenImport(string $raw): callable
{
    $key = 'projects/demo';
    return static function () use ($key, $raw): string {
        $key .= $raw;
        return <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($key)</weak_warning>;
    };
}
