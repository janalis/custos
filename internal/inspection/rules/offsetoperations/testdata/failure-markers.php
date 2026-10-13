<?php
/**
 * @return array|bool|null
 */
function fetch_row(string $sql) { return lookup($sql); }

/** @param string|bool $s */
function markers(int $id, $s) {
    // A lookup returning a row or false: the false is a failure marker.
    $row = fetch_row('SELECT ' . $id);
    echo $row['name'];
    echo $s[0];
    // Booleans are valid array keys (cast to 0/1).
    $byFlag = [false => 'no', true => 'yes'];
    echo $byFlag[$id > 0];
    $count = 3;
    echo <error descr="'$count' does not support offset access (types: int).">$count[0]</error>;
}

function defaults($tables = false) {
    // The default is only one of the values a caller may pass.
    if (isset($tables['product'])) {
        $tables['product_lang'] = true;
    }
    return $tables;
}

function loops(array $rows) {
    // The loop value is unknown: `false` is not the only value.
    $last = false;
    foreach ($rows['actions'] as $action) {
        $last = $action;
    }
    return $last['url'];
}
