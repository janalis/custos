<?php
$multiple = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', '$11 and $12', 'r')</warning>;
$double = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', "\$11", 'r')</warning>;
$unescaped = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', "$11", 'r')</warning>;
$quoted = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', 'it\'s $11\n', 'r')</warning>;
$escaped = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', '\\$11 and $12', 'r')</warning>;
$encoded = <warning descr="Brace the replacement backreference before a literal digit.">preg_replace('/(r)/', "\x2411", 'r')</warning>;
