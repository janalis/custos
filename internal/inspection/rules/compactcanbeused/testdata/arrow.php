<?php
// compact() inside an arrow function sees only the variables its body
// names: captured outer variables would be lost.
function rows($id, $tag) {
    $lazy = fn () => ['id' => $id, 'tag' => $tag];
    $mixed = fn ($id) => ['id' => $id, 'tag' => $tag];
    $own = fn ($id, $tag) => <weak_warning descr="Replace with 'compact('id', 'tag')'.">[</weak_warning>'id' => $id, 'tag' => $tag];
    return [$lazy, $mixed, $own];
}
