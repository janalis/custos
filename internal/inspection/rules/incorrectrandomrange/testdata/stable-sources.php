<?php
// Bounds whose every binding is visible: still reported.
function stableLocals(array $xs, $seed = 20)
{
    $floor = 100;
    $copy = 9;
    $copy = $copy;
    $label = abs($floor) . $xs[$floor];
    $reader = fn () => $floor + 1;
    $logger = function () use ($floor) { return $floor; };
    $pair = [$floor, 'floor' => $floor];
    foreach ($xs as $x) {
        $label .= $x;
    }
    return [
        <error descr="Minimum is greater than maximum in this random range.">mt_rand($floor, 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">mt_rand($copy, 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">mt_rand(($seed), 1)</error>,
        <error descr="Minimum is greater than maximum in this random range.">mt_rand($floor ?? $floor, 1)</error>,
    ];
}

class Wheel
{
    private $slots;
    private $spins = 50;
    private static $stops = 30;
    private $first = 7, $last = 7;

    public function __construct()
    {
        $this->slots = 0;
        $this->spins = $this->spins;
    }

    public function spin($name)
    {
        return [
            <error descr="Minimum is greater than maximum in this random range.">random_int(1, $this->slots)</error>,
            <error descr="Minimum is greater than maximum in this random range.">random_int($this->spins, 1)</error>,
            <error descr="Minimum is greater than maximum in this random range.">random_int(self::$stops, 1)</error>,
            <error descr="Minimum is greater than maximum in this random range.">random_int(static::$stops, 1)</error>,
            <error descr="Minimum is greater than maximum in this random range.">random_int($this->last, 1)</error>,
            $this->$name,
            self::$$name,
        ];
    }
}

final class SealedSampler
{
    const MAX = 0;

    public function pick()
    {
        return <error descr="Minimum is greater than maximum in this random range.">mt_rand(1, static::MAX)</error>;
    }
}

class PinnedSampler
{
    const MIN = 1;
    final const MAX = 0;

    public function pick()
    {
        return <error descr="Minimum is greater than maximum in this random range.">mt_rand(1, static::MAX)</error>;
    }
}

function anonymousSampler()
{
    return new class {
        const MAX = 0;

        public function pick()
        {
            return <error descr="Minimum is greater than maximum in this random range.">mt_rand(1, static::MAX)</error>;
        }
    };
}
