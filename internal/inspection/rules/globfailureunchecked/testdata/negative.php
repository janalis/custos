<?php
function good($pattern) { $paths = glob($pattern); if ($paths === false) { return; } foreach ($paths as $path) { echo $path; } }
