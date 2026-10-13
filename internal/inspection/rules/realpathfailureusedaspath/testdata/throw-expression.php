<?php
function thrower(string $dir) { $root = realpath($dir); $root !== false || throw new RuntimeException('missing'); return is_file($root . '/a.json'); }
