<?php
function payload($id, $tag, $note) {
    $a = compact('id', 'tag');
    $b = compact("id", 'note', 'tag');
    send(compact('note', 'id'));
    $c = compact('id', 'tag');
    return [$a, $b, $c];
}
