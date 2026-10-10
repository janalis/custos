<?php
$m = fileperms($path); if ($m !== false && ($m & 0002) === 0) { return true; }
