<?php
$root = realpath($path); if ($root === false) { throw new RuntimeException(); } file_get_contents($root . '/config.json');
