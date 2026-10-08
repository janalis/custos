<?php
function bodyDemo($s, $m) {
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u).">'/v[0-9]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[:digit:]' as '\d' (same result without /u).">'/[[:digit:]]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\d]' as '\D' (same result).">'/[^\d]+/u'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[A-Za-z0-9_]' as '\w' (same result without /u).">'/[a-zA-Z0-9_]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^A-Za-z0-9_]' as '\W' (same result without /u).">'/[^0-9A-Za-z_]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\s]' as '\S' (same result).">'/[^\s]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^\w]' as '\W' (same result).">'/[^\w]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[:word:]' as '\w' (same result without /u).">'/[[:word:]]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[^0-9]' as '\D' (same result without /u).">'/[^0-9]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u)."><weak_warning descr="Write '[^\s]' as '\S' (same result).">'/[0-9]-[^\s]/'</weak_warning></weak_warning>, $s, $m);
    preg_match('/[0-9,]/', $s, $m);

    preg_match(<weak_warning descr="Collapse '[ab][ab]' into '[ab]' with a counted quantifier.">'/[ab][ab]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab]{2}[ab]+' into '[ab]' with a counted quantifier.">'/x[ab]{2}[ab]+/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab][ab]' into '[ab]' with a counted quantifier.">'/[ab][ab][ab]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab][ab][ab][ab]' into '[ab]' with a counted quantifier.">'/[ab][ab][ab][ab]/'</weak_warning>, $s, $m);
    preg_match('/[ab]x[ab]/', $s, $m);
    preg_match('/[ab][ba]/', $s, $m);

    preg_match(<error descr="Class [\w\d.] is redundant: \d is already covered by \w.">'/[\w\d.]/'</error>, $s, $m);
    preg_match(<error descr="Class [-\W\D] is redundant: \D is already covered by \W.">'/[-\W\D]/'</error>, $s, $m);
    preg_match(<error descr="Class [\w\d] is redundant: \d is already covered by \w."><error descr="Class [\W\D] is redundant: \D is already covered by \W.">'/[\w\d][\W\D]/'</error></error>, $s, $m);
    preg_match('/[\d\s]/', $s, $m);
    preg_match('/[\w\D]/', $s, $m);

    preg_match(<error descr="Nested quantifier (\w+)+ risks catastrophic backtracking.">'/(\w+)+$/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\s*)+ risks catastrophic backtracking.">'/(?:\s*)+x/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\d+)* risks catastrophic backtracking.">'/(a|\d+)*z/'</error>, $s, $m);
    preg_match(<error descr="Nested quantifier (\s+)+ risks catastrophic backtracking.">'/k(\s+|-(?=\d))+/'</error>, $s, $m);
    preg_match('/(\d+){2}/', $s, $m);
    preg_match('/(\d+)*+/', $s, $m);
    preg_match('/(?>(\s+|x)+)/', $s, $m);
    preg_match('/(\d{1,})+/', $s, $m);

    preg_match(<weak_warning descr="Tag content matched with '.' likely needs the /s flag.">'#<li>.*</li>#'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Tag content matched with '.' likely needs the /s flag.">'#<li>.+?</li>#'</weak_warning>, $s, $m);
    preg_match('#<li>.*</li>#s', $s, $m);
    preg_match('#<li>[^<]*</li>#', $s, $m);

    preg_match('/café/', $s, $m);
    preg_match('/café/u', $s, $m);
    preg_replace('/\[entité\]/', 'x', $s);
    preg_match(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/[éè]/'</error>, $s, $m);
    preg_match(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/[^\]é]/'</error>, $s, $m);
    preg_match(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/né+/'</error>, $s, $m);
    preg_match(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/é/i'</error>, $s, $m);
    preg_match('/€+/u', $s, $m);
    preg_match(<error descr="Unicode escapes (\p, \P, \X) need the /u flag.">'/\pL+/'</error>, $s, $m);
    preg_match(<error descr="Unicode escapes (\p, \P, \X) need the /u flag.">'/k\X/'</error>, $s, $m);
    preg_match('/\\p/', $s, $m);
    preg_quote('/ü/', '/');
}
