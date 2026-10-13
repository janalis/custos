<?php
namespace App\View;

function parse_str($query)
{
    return $query;
}

class Page
{
    public static $slot;

    public function viaExtract(array $data)
    {
        $early = compact(<error descr="Variable '$title' may be undefined when compact() runs.">'title'</error>);
        \extract($data);
        return $early + compact('title', 'body');
    }

    public function viaVariableVariable(array $data)
    {
        foreach ($data as $key => $value) {
            $$key = $value;
        }
        return compact('title');
    }

    public function viaBraces()
    {
        ${'title'} = 1;
        return compact('title');
    }

    public function viaInclude()
    {
        require __DIR__ . '/vars.inc';
        return compact('title');
    }

    public function viaEval($code)
    {
        eval($code);
        return compact('title');
    }

    public function viaLegacyParseStr($query)
    {
        \parse_str($query);
        return compact('title');
    }

    public function viaArrow(array $data)
    {
        $load = fn() => extract($data);
        return compact('title');
    }

    public function parseStrWithResult($query)
    {
        \parse_str($query, $out);
        return compact('out', <error descr="Variable '$title' may be undefined when compact() runs.">'title'</error>);
    }

    public function namespacedParseStr($query)
    {
        parse_str($query);
        return compact(<error descr="Variable '$title' may be undefined when compact() runs.">'title'</error>);
    }

    public function staticProperty($name)
    {
        self::$$name = 1;
        return compact(<error descr="Variable '$title' may be undefined when compact() runs.">'title'</error>);
    }

    public function nestedScope(array $data)
    {
        $load = function () use ($data) {
            extract($data);
        };
        return compact(<error descr="Variable '$title' may be undefined when compact() runs.">'title'</error>);
    }

    public function self()
    {
        $bound = function () {
            return fn() => compact('this');
        };
        $unbound = static function () {
            return compact(<error descr="Variable '$this' may be undefined when compact() runs.">'this'</error>);
        };
        $arrow = static fn() => compact(<error descr="Variable '$this' may be undefined when compact() runs.">'this'</error>);
        return [compact('this'), $bound, $unbound, $arrow];
    }

    public static function make()
    {
        return compact(<error descr="Variable '$this' may be undefined when compact() runs.">'this'</error>);
    }
}

function plain()
{
    return compact(<error descr="Variable '$this' may be undefined when compact() runs.">'this'</error>);
}

$free = function () {
    return compact(<error descr="Variable '$this' may be undefined when compact() runs.">'this'</error>);
};
