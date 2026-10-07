<?php
function payload($id, $tag, $note) {
    $a = <weak_warning descr="Replace with 'compact('id', 'tag')'.">[</weak_warning>'id' => $id, 'tag' => $tag];
    $b = <weak_warning descr="Replace with 'compact(&quot;id&quot;, 'note', 'tag')'.">array</weak_warning>("id" => $id, 'note' => $note, 'tag' => $tag,);
    send(<weak_warning descr="Replace with 'compact('note', 'id')'.">[</weak_warning>'note' => $note, 'id' => $id]);
    $c = <weak_warning descr="Replace with 'compact('id', 'tag')'.">ARRAY</weak_warning>('id' => $id, /* c */ 'tag' => $tag);
    return [$a, $b, $c];
}
