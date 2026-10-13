<?php
function bad($path) { $resolved = realpath($path); if ($resolved === false) { return false; } if (<warning descr="Require a directory boundary in the containment check.">str_starts_with($resolved, '/srv/files')</warning>) { return true; } return false; }
