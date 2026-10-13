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

    public function <weak_warning descr="Declare ': int' as the return type.">raisedLiteral</weak_warning>(int $base)
    {
        $result = $base;
        $result **= 2;
        $result += 1;
        $this->total = $base ** 2;
        return $result;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">square</weak_warning>(int $n)
    {
        return $n ** 2;
    }

    public function <weak_warning descr="Declare ': float' as the return type.">floating</weak_warning>(float $base, int $exp)
    {
        return $base ** $exp;
    }

    public function <weak_warning descr="Declare ': float' as the return type.">root</weak_warning>(int $n)
    {
        return $n ** 0.5;
    }

    public function <weak_warning descr="Declare ': int' as the return type.">scale</weak_warning>(int $n)
    {
        return $n * 4;
    }
}
