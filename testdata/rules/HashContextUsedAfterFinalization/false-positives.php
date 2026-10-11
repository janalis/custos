<?php

function fresh(): array
{
    return [hash_init('sha256'), 1];
}

function shortDestructuring(string $a): string
{
    $h = hash_init('sha256');
    $first = hash_final($h);
    [$h, $label] = [hash_init('sha256'), 'second'];
    hash_update($h, $a);
    return $first . $label . hash_final($h);
}

function listDestructuring($h): void
{
    hash_final($h);
    list($h, $count) = fresh();
    hash_update($h, 'x');
}

function keyedDestructuring($h, array $source): void
{
    hash_final($h);
    ['context' => $h] = $source;
    hash_update($h, 'x');
}

function foreachValue($h, array $contexts): void
{
    hash_final($h);
    foreach ($contexts as $h) {
        hash_update($h, 'x');
    }
}

function foreachKey($h, array $map): void
{
    hash_final($h);
    foreach ($map as $h => $unused) {
        hash_update($h, 'x');
    }
}

function caught($h): void
{
    hash_final($h);
    try {
        risky();
    } catch (\Exception $h) {
        hash_update($h, 'x');
    }
}

function extracted(array $vars): void
{
    $h = hash_init('sha256');
    hash_final($h);
    extract($vars);
    hash_update($h, 'x');
}

function renew(): void
{
    $GLOBALS['ctx'] = hash_init('sha256');
}

function globalContext(string $next): string
{
    global $ctx;
    $done = hash_final($ctx);
    renew();
    hash_update($ctx, $next);
    return $done;
}

function staticContext(int $depth): void
{
    static $h;
    hash_final($h);
    Pool::refill($depth);
    hash_update($h, 'x');
}

function referenceParameter(&$h, $pool): void
{
    hash_final($h);
    $pool->refill();
    hash_update($h, 'x');
}

function referenceParameterConstructed(&$h): void
{
    hash_final($h);
    new Refill;
    hash_update($h, 'x');
}

function passedToMethod($h, $pool): void
{
    hash_final($h);
    $pool->renew($h);
    hash_update($h, 'x');
}

function passedToStaticMethod($h): void
{
    hash_final($h);
    Pool::renew($h);
    hash_update($h, 'x');
}

function passedToConstructor($h): void
{
    hash_final($h);
    new Renewal($h);
    hash_update($h, 'x');
}

function matchArms($h, int $op, string $data): void
{
    match ($op) {
        1 => print(hash_final($h) . "\n"),
        default => hash_update($h, $data),
    };
}

function matchResult($h, int $op): void
{
    $digest = match ($op) {
        1 => hash_final($h),
        default => '',
    };
    hash_update($h, $digest);
}

function nullSafeArgument($h, ?Sink $sink): void
{
    $sink?->write(hash_final($h));
    hash_update($h, 'x');
}

function nullSafeChain($h, ?Sink $sink): void
{
    $sink?->inner->target()->write(hash_final($h));
    hash_update($h, 'x');
}

function propertyContext(object $o, int $op): void
{
    match ($op) {
        1 => hash_final($o->h),
        default => hash_update($o->h, 'x'),
    };
}

function replaced(): void
{
    $h = hash_init('sha256');
    hash_final($h);
    $h = hash_init('sha256');
    hash_update($h, 'x');
    $later = function () use ($h) {
        hash_final($h);
    };
    hash_update($h, 'y');
    strlen('x');
    hash_final();
}

function otherContextFinalized($other): void
{
    $h = hash_init('sha256');
    hash_final($h);
    [$h, $label] = [hash_init('sha256'), 'second'];
    hash_final($other);
    hash_update($h, 'x');
}
