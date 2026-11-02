<?php
// Null written to a reference parameter empties the caller's variable.
final class Job
{
    private static function start(array &$options, array $local, array &$other, \Closure $f): void
    {
        echo count($options);
        $options = null;
        $local = <warning descr="Assigning a value of type null does not match the parameter's declared type.">null</warning>;
        $other = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'done'</warning>;
        $other .= 'x';
    }
}
