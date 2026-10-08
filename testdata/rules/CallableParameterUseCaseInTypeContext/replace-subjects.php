<?php
// A replacement on a subject of unknown type returns a string for a string
// and an array for an array: no report.
/** @param string $where */
function where($where, $raw, array $list) {
    $where = <warning descr="Assigning a value of type array does not match the parameter's declared type.">str_replace('a', 'b', $list)</warning>;
    require 'filters.php';
    $where = preg_replace('/a/', 'b', $where);
    $where = substr_replace($raw, 'x', 0, 1);
    $where = preg_replace_callback_array([], $raw);
    $where = <warning descr="Assigning a value of type array does not match the parameter's declared type.">str_replace('a', 'b')</warning>;
    $where = strtolower($raw);
    return $where;
}
