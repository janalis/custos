<?php
function render(array $rows, array $cfg, array $state, \SplStack $jobs)
{
    foreach ($rows as $row) {
        <weak_warning descr="Statement does not depend on the loop; move it out.">store($cfg['title']);</weak_warning>
        <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($cfg['debug']['enabled']) {
            trace($cfg['debug']);
        }
        emit($row);
    }

    foreach ($rows as $row) {
        $state['seen']++;
        echo $state['seen'];            // written above: connected
    }

    foreach ($rows as $row) {
        $state['total'][$row] = 1;
        log_total($state['total']);     // nested element write: connected
    }

    foreach ($rows as $row) {
        sort($cfg['order']);
        apply($cfg['order']);           // by-reference argument: connected
    }

    foreach ($rows as $row) {
        unset($state['tmp']);
        echo $state['tmp'];             // unset: connected
    }

    foreach ($rows as $row) {
        $alias = &$state['ref'];
        echo $state['ref'];             // reference binding: connected
    }

    foreach ($rows as $row) {
        $state['queue']->push($row);
        echo $state['queue']->count();  // method call on the element: connected
    }
}
