<?php
class Mailer {
    public static function send($to, $body) { return true; }
}
$m = new Mailer();

$hook($to, $log);
Mailer::send($to, $body);
$m->send($to, $body);
$m::send(...$rest);
$m->$verb();
$m->{"on{$evt}"}();
App\Mailer::send($to);
Other::make();
call_user_func($hook, $to, $body);
forward_static_call('Mailer::send', 1, 2);
