<?php
function write_rules($fd, array $hosts, bool $fixed)
{
    foreach ($hosts as $host) {
        fwrite($fd, '# rules' . PHP_EOL); // one header per host
        fputs($fd, $host);
        if ($fixed) {
            $weight = 50;
        } else {
            $weight = mt_rand(50, 90); // a new value on every iteration
        }
        $stats[$host] = $weight;
        <weak_warning descr="Statement does not depend on the loop; move it out.">if</weak_warning> ($fixed) {
            $mode = 'fixed';
        }
        $modes[$host] = $mode;
    }
    return [$stats, $modes];
}
