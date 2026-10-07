<?php
function lastCrumb(string $path): string {
    $crumbs = [['home', '/']];
    foreach (explode('/', $path) as $part) {
        $crumbs[] = [$part, $path];
    }
    [$label] = array_pop($crumbs);
    [$first] = array_shift($crumbs);
    [$tail] = end($crumbs);              // E8: array element or false
    return $label . $first . $tail;
}
