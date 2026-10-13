<?php

function literalCount()
{
    $width = -2;
    return <error descr="Supply a nonnegative array fill count.">array_fill(0, $width, 'c')</error>;
}

function unrelatedCounter(array $rows)
{
    $width = -2;
    $seen = 0;
    foreach ($rows as $row) {
        $seen++;
        $rows[$seen]--;
    }
    $copy = function () use ($width) {
        return $width;
    };
    $pair = [$seen, &$rows];
    return <error descr="Supply a nonnegative array fill count.">array_fill(0, $width, $copy)</error>;
}

function copiedLiteral()
{
    $seed = -3;
    $width = $seed;
    $width = $seed;
    $other = $width + $width;
    return <error descr="Supply a nonnegative array fill count.">array_fill(0, $width, $other)</error>;
}

function nestedScopesAreSeparate()
{
    $width = -1;
    $inner = function () {
        $width = 0;
        $width++;
        extract(['width' => 2]);
        return $width;
    };
    $arrow = fn($width) => $width++;
    function declaredInside()
    {
        $width = 0;
        $width++;
    }
    $object = new class {
        public function step()
        {
            $width = 0;
            $width++;
        }
    };
    return <error descr="Supply a nonnegative array fill count.">array_fill(0, $width, [$inner, $arrow, $object])</error>;
}

$fill = fn() => [$inArrow = -1, <error descr="Supply a nonnegative array fill count.">array_fill(0, $inArrow, 'c')</error>];
