<?php
function dispatch(array $events) {
    while ($e = array_shift($events)) {
        switch ($e->type) {
            case 'skip':
                <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
            case 'retry':
                if ($e->tries > 3) {
                    <error descr="Inside 'switch', 'continue' acts like 'break'; use 'continue 2' to reach the loop.">continue;</error>
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
