<?php
function payload($id, $tag, $note, $rows) {
    $c = ['id' => $id];
    $d = ['id' => $tag, 'tag' => $id];
    $e = ['id' => $id, 'tag' => trim($tag)];
    $f = ['id' => $id, 'tag' => $tag, $note];
    $g = ['ID' => $id, 'tag' => $tag];
    $h = [0 => $id, 1 => $tag];
    $i = [KEY => $id, 'tag' => $tag];
    $j = ['id' => ($id), 'tag' => $tag];
    $k = ['id' => &$id, 'tag' => &$tag];
    $l = ['id' => $id, ...$rows];
    $m = [];
    ['id' => $id, 'tag' => $tag] = load();
    [['id' => $id, 'tag' => $tag]] = load();
    foreach ($rows as ['id' => $id, 'tag' => $tag]) {}
    return [$c, $d, $e, $f, $g, $h, $i, $j, $k, $l, $m];
}
