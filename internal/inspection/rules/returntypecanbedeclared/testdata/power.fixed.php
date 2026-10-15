<?php
final class Calc
{
    // A negative exponent gives a float.
    public function power(int $base, int $exp)
    {
        return $base ** $exp;
    }

    public function powerPlus(int $base, int $exp)
    {
        return 1 + $base ** $exp;
    }

    public function negative(int $base)
    {
        return $base ** -1;
    }

    public function assigned(int $base, int $exp)
    {
        $result = $base ** $exp;
        return $result;
    }

    public function raised(int $base, int $exp)
    {
        $result = $base;
        $result **= $exp;
        return $result;
    }

    public function raisedLiteral(int $base)
    {
        $result = $base;
        $result **= 2;
        $result += 1;
        $this->total = $base ** 2;
        return $result;
    }

    public function square(int $n)
    {
        return $n ** 2;
    }

    public function floating(float $base, int $exp)
    {
        return $base ** $exp;
    }

    public function root(int $n)
    {
        return $n ** 0.5;
    }

    public function literalPower(): int
    {
        return 2 ** 3;
    }

    public function overflowingPower()
    {
        return 2 ** 63;
    }

    public function scale(int $n): int
    {
        return $n * 4;
    }
}
