<?php
function slug($title)
{
    $title = str_replace('&', /* why */ 'and', $title);
    $title = str_replace(' ', '-', $title);
    $nested = str_replace('a', 'b', str_replace('c', /* keep */ 'd', $title));
    $same = str_replace(['x', /* one */ 'x'], 'y', $title);
    return [$title, $nested, $same];
}
