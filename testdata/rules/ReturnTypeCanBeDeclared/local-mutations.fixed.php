<?php
final class MutationReturns
{
    public function incrementNull(): int
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

    public function capturedValue(): string
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
