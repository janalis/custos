<?php // presets: per
$options = ['local_cert' => isset($ssl['local_cert']) ? self::pathIdentity($ssl['local_cert'], $ssl['other'], $ssl['third']) : null];
unset(
    $a[1],
    $b
);
