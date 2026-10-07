<?php
class Gate
{
    protected function pass() {}
    public function open() {}
}

function pick(bool $alt, array $names)
{
    $cb = 'Gate::pass';
    if ($alt) {
        $cb .= 'Twice';
    }
    $pair = ['Gate', 'open'];
    foreach ($names as $n) {
        $pair += [2 => $n];
    }
    return [is_callable($cb), is_callable($alt ? $cb : 'trim'), is_callable($pair)];
}

function stable()
{
    $cb = 'Gate::pass';
    return is_callable(<warning descr="Method 'pass' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'pass' is not static but is referenced without an object.">$cb</warning></warning>);
}
