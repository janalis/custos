<?php
function good($privateDir, $secret) { $path = tempnam($privateDir, 'job'); if ($path === false) { return; } if (realpath(dirname($path)) !== realpath($privateDir)) { unlink($path); return; } file_put_contents($path, $secret); }
