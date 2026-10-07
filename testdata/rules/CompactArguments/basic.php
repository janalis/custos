<?php
function summary($net, $tax)
{
    $gross = $net + $tax;
    return compact(
        'net',
        "tax",
        'gross',
        'discount',
        <error descr="Variable '$discount' may be undefined when compact() runs.">'discount'</error>,
        <error descr="Variable '$$gross' may be undefined when compact() runs.">'$gross'</error>,
        "{$gross}",
        $net
    );
}

function lateAssignment($unit)
{
    $payload = compact('unit', <error descr="Variable '$count' may be undefined when compact() runs.">'count'</error>);
    $count = 3;
    return $payload;
}

class Report
{
    public function build()
    {
        $title = 'Q3';
        $render = function () use ($title) {
            return compact('title', <error descr="Variable '$footer' may be undefined when compact() runs.">'footer'</error>);
        };
        return compact('title', 'render');
    }
}

function arrow($a)
{
    $b = 1;
    return fn() => \compact('a', 'b', <error descr="Variable '$c' may be undefined when compact() runs.">'c'</error>);
}
