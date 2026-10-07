<?php
class Mailer {
    public static function send($to, $body) { return true; }
}
$m = new Mailer();

<weak_warning descr="Call it directly: '$hook($to, $log)'.">call_user_func($hook, $to, $log)</weak_warning>;
<weak_warning descr="Call it directly: 'Mailer::send($to, $body)'.">call_user_func(['Mailer', 'send'], $to, $body)</weak_warning>;
<weak_warning descr="Call it directly: '$m->send($to, $body)'.">call_user_func(array($m, 'send'), $to, $body)</weak_warning>;
<weak_warning descr="Call it directly: '$m::send(...$rest)'.">forward_static_call([$m, 'send'], ...$rest)</weak_warning>;
<weak_warning descr="Call it directly: '$m->$verb()'.">call_user_func([$m, $verb])</weak_warning>;
<weak_warning descr="Call it directly: '$m->{&quot;on{$evt}&quot;}()'.">call_user_func([$m, "on{$evt}"])</weak_warning>;
<weak_warning descr="Call it directly: 'App\Mailer::send($to)'.">\call_user_func('App\\Mailer::send', $to)</weak_warning>;
<weak_warning descr="Call it directly: 'Other::make()'.">call_user_func(['Base', 'Other::make'])</weak_warning>;
<weak_warning descr="Pass the arguments inline: 'call_user_func($hook, $to, $body)'.">call_user_func_array($hook, [0 => $to, 1 => $body])</weak_warning>;
<weak_warning descr="Pass the arguments inline: 'forward_static_call('Mailer::send', 1, 2)'.">\forward_static_call_array('Mailer::send', array(1, 2))</weak_warning>;
