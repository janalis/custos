<?php

namespace Paths {
    function dirname(string $p, int $levels = 1): string { return $p; }

    // Paths\dirname() is a user function: not collapsed.
    $mine = dirname(dirname($file));
    $root = \dirname($file, 2);
}
