<?php

namespace Build {
    function dirname(string $path): string { return $path; }

    // dirname() here is Build\dirname(), not the global function.
    $own = dirname(__FILE__);
    $global = __DIR__;
}
