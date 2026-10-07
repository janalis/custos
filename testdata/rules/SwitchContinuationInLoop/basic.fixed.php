<?php
function dispatch(array $events) {
    while ($e = array_shift($events)) {
        switch ($e->type) {
            case 'skip':
                continue 2;
            case 'retry':
                if ($e->tries > 3) {
                    continue 2;
                }
                continue 2;
            case 'batch':
                for ($i = 0; $i < 2; $i++) {
                    continue;
                }
                break;
            case 'later':
                $cb = function () { continue; };
                break;
        }
    }
}
