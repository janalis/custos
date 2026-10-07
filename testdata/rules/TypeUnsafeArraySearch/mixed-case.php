<?php
function g($id, array $ids) {
    return [
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">In_Array($id, $ids)</weak_warning>,
        <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">\ARRAY_SEARCH($id, $ids)</weak_warning>,
    ];
}
