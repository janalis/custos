<?php
// Before PHP 8 the argument array's keys are ignored.
<weak_warning descr="Pass the arguments inline: 'call_user_func($hook, $to, $body)'.">call_user_func_array($hook, ['to' => $to, $body])</weak_warning>;
