<?php
function slug($title) {
    $a = str_replace('-', '_', $title);
    $b = \str_replace('.', '/', $title);
    return $a . $b;
}
