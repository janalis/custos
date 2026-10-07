<?php
namespace App;

function in_array($needle, $haystack) { return false; }

function f($id, array $ids) {
    return [
        in_array($id, $ids),
        \App\in_array($id, $ids),
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\in_array($id, $ids)</weak_warning>,
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">array_search($id, $ids)</weak_warning>,
    ];
}
