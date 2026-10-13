<?php
class CopySource
{
    public $label;
}

function copyPositional(?CopySource $source)
{
    return clone(<warning descr="Possible null dereference.">$source</warning>, ['label' => 'copy']);
}

function copyNamed(?CopySource $source, ?string $label)
{
    return clone(withProperties: ['label' => $label], object: <warning descr="Possible null dereference.">$source</warning>);
}

function copyOverrides(CopySource $source, ?string $label)
{
    return clone($source, ['label' => $label]);
}

function cloneFactory(?CopySource $unused)
{
    return clone(...);
}
