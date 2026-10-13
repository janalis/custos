<?php
class Dispatcher
{
    public function render(array $displayData, string $path): void
    {
        extract($displayData);
        unset($displayData);
        include $path;
    }

    public function vars(array $data): array
    {
        unset($data);
        return get_defined_vars();
    }

    public function packed(array $data, $title): array
    {
        unset($data);
        return compact('title');
    }

    public function evaluated(string $code, $tmp): void
    {
        unset($tmp);
        eval($code);
    }

    public function plain($tmp): void
    {
        <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($tmp);</weak_warning>
        $f = function () { include 'x.php'; };
    }
}

function allArgs($first)
{
    unset($first);
    return func_get_args();
}

function oneArg($first)
{
    unset($first);
    return func_get_arg(0);
}

function dynamicName($value)
{
    unset($value);
    $name = 'value';
    return isset($$name);
}

function exposedEarly($value, $path)
{
    include $path;
    unset($value);
    return $path;
}

// A static property fetched by a computed name does not read local variables.
function staticByName($value, $name)
{
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($value);</weak_warning>
    return Registry::$$name;
}
