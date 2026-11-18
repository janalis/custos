<?php
$entries = scandir($directory); if ($entries !== false) { foreach ($entries as $entry) { echo $entry; } }
