<?php
final class Finder
{
    public function chain(string $path): string
    {
        // either operand of `?:` is a call: false marks a failure
        $path = realpath($path . '/a') ?: realpath($path . '/b') ?: realpath($path);
        return $path;
    }

    public function fallback(string $path): string
    {
        $path = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">realpath($path) ?: false</warning>;
        return (string) $path;
    }

    public function computed(string $path): string
    {
        $path = <warning descr="Assigning a value of type bool does not match the parameter's declared type.">$path . '/' ?: realpath($path)</warning>;
        return (string) $path;
    }
}
