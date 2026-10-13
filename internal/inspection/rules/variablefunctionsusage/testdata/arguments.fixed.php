<?php
$hook(to: $to, body: $body);
call_user_func(...$all);
call_user_func_array($hook, ...$lists);
call_user_func_array($hook, args: [$to]);
