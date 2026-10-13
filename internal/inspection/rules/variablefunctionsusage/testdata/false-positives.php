<?php
class Mailer {
    public static function send($to, $body) { return true; }
}
$m = new Mailer();

call_user_func_array($hook, []);
call_user_func_array($hook, [&$to]);
call_user_func_array($hook, [...$to]);
call_user_func_array($hook, $args);
call_user_func_array($hook, [1], 2);
call_user_func([$m, 'parent::send'], $to);
call_user_func([$pool[0], 'send'], $to);
call_user_func("{$cls}::send", $to);
call_user_func(build(), $to);
call_user_func([], $to);
call_user_func();
$name = 'mailer';
call_user_func([$name, 'send'], $to);
call_user_func([$unknown, 'send'], $to);
