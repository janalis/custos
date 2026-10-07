<?php
function demo($flag, $rows, $n) {
    if ($flag) {
        log_it();
    }
    elseif ($n > 3) {
        $n--;
    }
    else {
        $n = 0;
    }

    if ($flag) { tick(); }
    else if ($n) {
        tock();
    }

    foreach ($rows as $row) {
        emit($row);
    }
    for ($i = 0; $i < $n; $i++) {
        ;
    }
    while (more()) {
        step();
    }
    do {
        pull();
    } while (pending());
}
