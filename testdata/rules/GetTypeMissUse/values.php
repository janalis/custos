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
        return <warning descr="Use '!is_string($v)' instead.">gettype($v) !== $this->valueType</warning>;
    }
}

function local($v)
{
    $t = 'array';
    return <warning descr="Use 'is_array($v)' instead.">gettype($v) === $t</warning>;
}

function either($v, $c)
{
    $t = $c ? 'array' : make();
    return gettype($v) === $t;
}

function invalid($v, $c)
{
    $t = $c ? <error descr="gettype() never returns 'int'.">'int'</error> : make();
    return gettype($v) === $t;
}
