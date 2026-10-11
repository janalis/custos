<?php
// scanf-only conversions and size modifiers.
sscanf($s, '%i %s%n', $a, $b, $n);
sscanf($s, '%hd', $h);
sscanf($s, '%ld %Lf', $l, $f);
sscanf($s, '%D', $d);
sscanf($s, '%5c%3s', $c, $w);
sscanf($s, '%2$d %1$s', $w, $d);
// A set runs to its closing bracket: `%` and a leading `]` are members.
sscanf($s, '%[^%]%%%s', $x, $y);
sscanf($s, '%[%%]x', $x);
sscanf($s, '%[]a]', $x);
sscanf($s, '%[^]a]', $x);
// Suppressed conversions assign nothing.
sscanf($s, '%*[^,],%d', $d);
sscanf($s, '%*5d%s', $w);
sscanf($s, '%*n');
fscanf($fh, '%i %n', $a, $n);

<error descr="This call needs 5 argument(s) in total.">sscanf</error>($s, '%i %s%n', $a, $b);
<error descr="This call needs 3 argument(s) in total.">sscanf</error>($s, '%[^%]', $x, $y);
<error descr="This call needs 3 argument(s) in total.">sscanf</error>($s, '%*d %hd');
<error descr="This call needs 4 argument(s) in total.">fscanf</error>($fh, '%2$d %1$s', $w);

// Not part of the scanf grammar: printf flags, precision and conversions.
sscanf($s, <error descr="Malformed format string.">'%-5d'</error>, $d);
sscanf($s, <error descr="Malformed format string.">'%.2f'</error>, $f);
sscanf($s, <error descr="Malformed format string.">'%b'</error>, $d);
sscanf($s, <error descr="Malformed format string.">'%G'</error>, $f);
sscanf($s, <error descr="Malformed format string.">'%lld'</error>, $d);
sscanf($s, <error descr="Malformed format string.">'%*2$d'</error>, $d);
sscanf($s, <error descr="Malformed format string.">'%[a-z'</error>, $w);
sscanf($s, <error descr="Malformed format string.">'%[^'</error>, $w);
sscanf($s, <error descr="Malformed format string.">'%d %'</error>, $d);
sscanf($s, <error descr="Malformed format string.">'%d %5'</error>, $d);
