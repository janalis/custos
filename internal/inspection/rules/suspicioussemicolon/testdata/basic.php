<?php
function poll($queue, $ready, $late) {
    if ($ready) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    elseif ($late) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    else <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    {
        notify();
    }

    while ($queue->next()) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    do <error descr="This ';' is the entire body of the statement; probably unintended.">;</error> while ($queue->busy());
    for ($i = 3; $i > 0; $i--) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    foreach ($queue as $job) <error descr="This ';' is the entire body of the statement; probably unintended.">;</error>
    if ($late) { ; }
    declare(ticks=1);
    notify();;
}
