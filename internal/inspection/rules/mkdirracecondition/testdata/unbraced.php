<?php
// The thrown re-check is braced so it cannot capture the outer else.
function prepare($dir, $ready) {
    if ($ready)
        <error descr="mkdir() outcome is ignored; use 'if (!mkdir($dir) && !is_dir(...)) { ... }'.">mkdir($dir);</error>
    else
        echo 'later';
}
