<?php
function make() {
    if (!mkdir('/tmp/x') && !is_dir('/tmp/x')) { throw new \RuntimeException(sprintf('Directory "%s" was not created', '/tmp/x')); }
}
