<?php
$h = hash_init('sha256'); hash_final($h); <error descr="Create a new hash context after finalization.">hash_update($h, 'next')</error>;

function repeated(string $data): string
{
    $h = hash_init('sha256');
    hash_update($h, $data);
    $digest = hash_final($h);
    <error descr="Create a new hash context after finalization.">hash_update($h, $data)</error>;
    return $digest . <error descr="Create a new hash context after finalization.">hash_final($h)</error>;
}

function namedArguments($h): void
{
    hash_final(context: $h);
    <error descr="Create a new hash context after finalization.">hash_update(data: 'x', context: $h)</error>;
}

function firstIteration($h, int $count): void
{
    hash_final($h);
    while ($count--) {
        <error descr="Create a new hash context after finalization.">hash_update($h, 'x')</error>;
        $h = hash_init('md5');
    }
}

function knownCallsOnly(): void
{
    static $h;
    $h = hash_init('sha256');
    hash_final($h);
    strlen('unrelated');
    <error descr="Create a new hash context after finalization.">hash_copy($h)</error>;
}

function sameArm($h, int $mode): void
{
    match ($mode) {
        default => [hash_final($h), <error descr="Create a new hash context after finalization.">hash_update($h, 'x')</error>],
    };
}

function otherNames($h, array $pairs, $logger): void
{
    hash_final($h);
    [$first, $second] = $pairs;
    foreach ($pairs as $key => $value) {
        $logger->debug($key);
    }
    try {
        Registry::touch($first);
    } catch (\Throwable $error) {
        new \ArrayObject([$second]);
    }
    <error descr="Create a new hash context after finalization.">hash_update($h, 'x')</error>;
}

function plainChain($h, $sink): void
{
    $sink->inner->write(hash_final($h));
    <error descr="Create a new hash context after finalization.">hash_update($h, 'x')</error>;
}
