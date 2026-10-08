<?php
// The thrown re-check is braced so it cannot capture the outer else.
function prepare($dir, $ready) {
    if ($ready)
        { if (!mkdir($dir) && !is_dir($dir)) { throw new \RuntimeException(sprintf('Directory "%s" was not created', $dir)); } }
    else
        echo 'later';
}
