<?php
function fromQuery($data, $query)
{
    $level = 20;
    parse_str($query);
    return gzencode($data, $level);
}
