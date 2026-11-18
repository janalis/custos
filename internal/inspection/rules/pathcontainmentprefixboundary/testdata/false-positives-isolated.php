<?php
function good($path) { $resolved = realpath($path); if ($resolved === false) { return false; } return $resolved === '/srv/files' || str_starts_with($resolved, '/srv/files/'); }
