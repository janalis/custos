<?php
// Bounds that are rebound in a way value discovery does not read.
function fromLoop(array $limits)
{
    $max = 0;
    $out = [];
    foreach ($limits as $max) {
        $out[] = mt_rand(5, $max);
    }
    return $out;
}

function fromLoopKey(array $limits)
{
    $max = 0;
    foreach ($limits as [$min, $max]) {
    }
    return mt_rand(5, $max);
}

function fromList(array $range)
{
    $max = 0;
    list($min, $max) = $range;
    return mt_rand(5, $max);
}

function fromNestedPattern(array $range)
{
    $max = 0;
    [$min, [$step, $max]] = $range;
    return mt_rand(5, $max);
}

function fromOutParameter(string $spec)
{
    $max = 0;
    sscanf($spec, 'd%d', $max);
    return mt_rand(5, $max);
}

function fromUnknownCallee($loader)
{
    $max = 0;
    $loader->fill($max);
    return mt_rand(5, $max);
}

function fromUndeclaredFunction()
{
    $max = 0;
    fill_limit($max);
    return mt_rand(5, $max);
}

function fromCallTimeReference(array $stack)
{
    $max = 0;
    array_push($stack, [&$max]);
    return mt_rand(5, $max);
}

function fromParameter(int $limit)
{
    $max = 0;
    $max = $limit;
    return mt_rand(5, $max);
}

function fromReferenceParameter(&$max = 0)
{
    return mt_rand(5, $max);
}

function fromAlias()
{
    $max = 0;
    $alias = &$max;
    $alias = 100;
    return mt_rand(5, $max);
}

function fromClosure()
{
    $max = 0;
    $raise = function () use (&$max) { $max = 100; };
    $raise();
    return mt_rand(5, $max);
}

function fromCatch()
{
    $max = 0;
    try {
        load_limits();
    } catch (Exception $max) {
    }
    return mt_rand(5, $max);
}

function fromStatic()
{
    static $max = 0;
    $max = 0;
    return mt_rand(5, $max);
}

function fromGlobal()
{
    global $max;
    $max = 0;
    load_limits();
    return mt_rand(5, $max);
}

function fromUnset(array $pool)
{
    $max = 0;
    unset($max);
    return mt_rand(5, $max ?? count($pool));
}

function fromElementWrite()
{
    $max = 0;
    $max[] = 9;
    return mt_rand(5, $max);
}

function fromExtract(array $settings)
{
    $max = 0;
    extract($settings);
    return mt_rand(5, $max);
}

function fromVariableVariable(string $name)
{
    $max = 0;
    $$name = 100;
    return mt_rand(5, $max);
}

function fromInclude(string $file)
{
    $max = 0;
    include $file;
    return mt_rand(5, $max);
}

function fromEval(string $code)
{
    $max = 0;
    eval($code);
    return mt_rand(5, $max);
}

function fromFallback($limit, $obj)
{
    return [mt_rand(5, $limit ?? 0), mt_rand(5, $limit ?: 0), mt_rand(5, $obj::MAX ?? 0)];
}

class Dice
{
    private $sides = 0;
    private $faces = 0;
    private $rolls = 0;
    private static $throws = 0;
    public $weight = 0;
    private $bias = 0 { get => $this->bias + 100; }

    public function setSides(int $n): void
    {
        $this->sides = $n;
        $this->faces = 6;
        $this->rolls++;
        self::$throws += 1;
    }

    public function roll(Dice $other)
    {
        $other->faces = 6;
        return [
            random_int(1, $this->sides),
            random_int(1, $this->faces),
            random_int(1, $this->rolls),
            random_int(1, self::$throws),
            random_int(1, $this->weight),
            random_int(1, $this->bias),
            random_int(1, $other->faces),
            random_int(1, Dice::$throws),
        ];
    }
}

trait SetsLimit
{
    public function widen() { $this->limit = 100; }
}

class TraitDice
{
    use SetsLimit;

    private $limit = 0;

    public function roll() { return random_int(1, $this->limit); }
}

abstract class Sampler
{
    const MAX = 0;

    public function pick($name)
    {
        return [mt_rand(1, static::MAX), mt_rand(1, $this::MAX), mt_rand(1, static::{$name} ?? 0)];
    }
}

final class TenSampler extends Sampler
{
    const MAX = 10;
}

mt_rand(5, $unset ?? 0);
