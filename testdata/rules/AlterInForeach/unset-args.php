<?php
function cleanup(array $rows, array $cache, $obj) {
    foreach ($rows as $row) {}
    unset($cache['rows'], <weak_warning descr="'$row' is not a reference here; unsetting it is unnecessary.">$row</weak_warning>);

    foreach ($rows as $id => $line) {}
    unset($obj->line, $cache[$id], <weak_warning descr="'$line' is not a reference here; unsetting it is unnecessary.">$line</weak_warning>);

    foreach ($rows as $entry) {}
    unset($tmp, $obj->entry);

    foreach ($rows as $cell) {}
    unset($cache['cell']);
}
