<?php
// Rewrites that would change the call or produce invalid PHP: no report.
function hooks($hook, $to, $log, $key, array $pool, $i) {
    call_user_func($hook, $to, &$log);
    call_user_func([$hook, 'run'], & $log);
    call_user_func_array($hook, ['to' => $to, $log]);
    call_user_func_array($hook, [$key => $to]);
    call_user_func([$pool[$i], 'Base::make'], $to);
    call_user_func([make(), 'Base::make']);
}
