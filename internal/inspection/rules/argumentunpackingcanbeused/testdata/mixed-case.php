<?php

function forward(array $parts): string
{
    return <warning descr="Call 'implode(...$parts)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">Call_User_Func_Array('implode', $parts)</warning>;
}
