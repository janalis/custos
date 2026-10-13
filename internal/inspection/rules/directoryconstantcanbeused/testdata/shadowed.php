<?php

namespace Build {
    function dirname(string $path): string { return $path; }

    // dirname() here is Build\dirname(), not the global function.
    $own = dirname(__FILE__);
    $global = <warning descr="Replace dirname(__FILE__) with __DIR__.">\DirName(__FILE__)</warning>;
}
