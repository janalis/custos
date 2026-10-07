<?php
namespace App;

function in_array($needle, $haystack) { return false; }

function f($id, array $ids) {
    return [
        in_array($id, $ids),
        \App\in_array($id, $ids),
        \in_array($id, $ids, true),
        array_search($id, $ids, true),
    ];
}
