<?php
function g($id, array $ids) {
    return [
        In_Array($id, $ids, true),
        \ARRAY_SEARCH($id, $ids, true),
    ];
}
