<?php
function pickMode(array $modes) {
    <warning descr="Loop body exits on the first iteration; the loop never repeats.">foreach</warning> ($modes as $mode) {
        switch ($mode) {
            case 'skip':
                continue;      // leaves the switch only
        }
        return $mode;
    }
    foreach ($modes as $mode) {
        switch ($mode) {
            case 'skip':
                continue 2;    // continues the foreach
        }
        return $mode;
    }
    return null;
}
