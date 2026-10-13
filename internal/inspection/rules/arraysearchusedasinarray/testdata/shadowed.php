<?php
namespace Inventory;

function array_search($needle, array $haystack) { return 0; }

function stocked(string $sku, array $items): bool {
    if (array_search($sku, $items)) {
        return true;
    }
    return Array_Search($sku, $items) === false || \Lookup\array_search($sku, $items) === false;
}

function listed(string $sku, array $items): bool {
    return <warning descr="Use 'in_array(...)' to test membership.">\array_search($sku, $items) !== false</warning>;
}
