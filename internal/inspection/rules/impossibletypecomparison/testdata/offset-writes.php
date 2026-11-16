<?php
namespace Shop;

// A write into false makes it an array (deprecated from PHP 8.1): the
// comparison may hold.
function vivify(): bool
{
    $list = false;
    $list[] = 1;
    return $list === [1];
}
