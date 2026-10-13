<?php
// A non-literal operand needs one complete value.
class NumericDef
{
    protected $valueType = 'integer';

    public function check($v)
    {
        return gettype($v) !== $this->valueType; // a subclass may redeclare it as 'double'
    }
}

class FloatDef extends NumericDef
{
    protected $valueType = 'double';
}

final class Fixed
{
    protected $valueType = 'string';

    public function check($v)
    {
        return !is_string($v);
    }
}

function local($v)
{
    $t = 'array';
    return is_array($v);
}

function either($v, $c)
{
    $t = $c ? 'array' : make();
    return gettype($v) === $t;
}

function invalid($v, $c)
{
    $t = $c ? 'int' : make();
    return gettype($v) === $t;
}
