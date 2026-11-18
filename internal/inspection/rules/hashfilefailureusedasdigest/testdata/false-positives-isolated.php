<?php
$h = hash_file('sha256', $path); if ($h !== false) { file_put_contents($manifest, $h); }
