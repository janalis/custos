<?php
final class MutationReturns
{
    public function <weak_warning descr="Declare ': int' as the return type.">incrementNull</weak_warning>()
    {
        $value = null;
        ++$value;
        return $value;
    }

    public function removed()
    {
        $value = 'before';
        unset($value);
        return $value;
    }

    public function <weak_warning descr="Declare ': string' as the return type.">capturedValue</weak_warning>()
    {
        $value = 1;
        $value = 'captured';
        $read = function () use ($value) { return $value; };
        return $read();
    }

    public function unknownCapture(array $values)
    {
        $value = 1;
        extract($values);
        $read = fn() => $value;
        return $read();
    }
}
