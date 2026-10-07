<?php
foreach (['a.php', 'b.php'] as $moduleFile) {
    require_once __DIR__ . '/' . $moduleFile;
}
unset($moduleFile);

function scoped(array $files) {
    foreach ($files as $file) {
        echo $file;
    }
    unset(<weak_warning descr="'$file' is not a reference here; unsetting it is unnecessary.">$file</weak_warning>);
}
