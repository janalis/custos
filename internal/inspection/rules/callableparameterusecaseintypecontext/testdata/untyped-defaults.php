<?php
function where_clause($table, $extra = false)
{
    // No declared or documented type: the default is only one value.
    if ($extra) {
        $extra = sprintf($extra, $table);
    }
    return $extra;
}

/** @param bool $flag */
function documented($flag = false)
{
    $flag = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'yes'</warning>;
    return $flag;
}

function declared(bool $flag = false)
{
    $flag = <warning descr="Assigning a value of type string does not match the parameter's declared type.">'yes'</warning>;
    return $flag;
}

/** @param null $unused */
function documented_null($unused = null)
{
    $unused = 'set';
    return $unused;
}
