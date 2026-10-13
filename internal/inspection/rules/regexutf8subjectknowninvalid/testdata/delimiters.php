<?php
// Bracket-style delimiters: the modifiers follow the matching closing bracket.
<error descr="Supply valid UTF-8 to a Unicode pattern.">preg_match('(a(b))u', "ab\xFF")</error>;
<error descr="Supply valid UTF-8 to a Unicode pattern.">preg_match('{a{1}}iu', "a\xFF")</error>;
<error descr="Supply valid UTF-8 to a Unicode pattern.">preg_split('[,]u', "a,\xFF")</error>;
// A "u" inside the body is not a modifier.
preg_match('[a[u]]', "u\xFF");
preg_match('{a{1}u}', "au\xFF");
preg_match('(a(u))', "au\xFF");
preg_match('<a<u>>', "au\xFF");
// Text after the first unescaped delimiter that is not a modifier list.
preg_match('~a~x~u', "a\xFF");
preg_match('~a~uq', "a\xFF");
