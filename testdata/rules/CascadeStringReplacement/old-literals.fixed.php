<?php
function literals($s)
{
    $s = str_replace(array('a', 'b', 'c'), array('1', '2', '3'), $s);
    return $s;
}
