<?php
class ReviewMatcher
{
    const RE = '/tmp/';
    protected $pattern = '/tmp/';

    public function run(string $s)
    {
        // a property or a late-bound constant may hold another pattern
        $r[] = preg_match($this->pattern, $s) ? 1 : 0;
        $r[] = preg_match(static::RE, $s) ? 1 : 0;
        $r[] = <warning descr="Replace with 'false !== strpos($s, &quot;tmp&quot;)'.">preg_match(self::RE, $s)</warning> ? 1 : 0;
        $r[] = <warning descr="Replace with 'false !== strpos($s, &quot;tmp&quot;)'.">preg_match((ReviewMatcher::RE), $s)</warning> ? 1 : 0;
        return $r;
    }
}

function reviewDiscovery(string $s, array $extra, string $re = '/tmp/')
{
    // the default is one possible value only
    $r[] = preg_match($re, $s) ? 1 : 0;
    $sep = '/,/';
    foreach ($extra as $sep) {
        $r[] = preg_split($sep, $s);
    }
    // the literal checks still follow the variable
    $old = <error descr="The /e flag was removed from PCRE; use a callback replacement.">'/tmp/e'</error>;
    $r[] = preg_match($old, $s) ? 1 : 0;
    return $r;
}

function reviewSubject(string $s, $v, array $q, bool $flag)
{
    $r[] = <warning descr="Replace with '&quot;a&quot; === $s'.">preg_match('/^a$/D', $s)</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with '&quot;a&quot; === $s . 'a''.">preg_match('/^a$/D', $s . 'a')</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with '&quot;a&quot; === ($flag ? 'zzz' : 'a')'.">preg_match('/^a$/D', $flag ? 'zzz' : 'a')</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with '&quot;a&quot; === (string) ($q['k'] ?? 'a')'.">preg_match('/^a$/D', $q['k'] ?? 'a')</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with '&quot;1&quot; === (string) $v'.">preg_match('/^1$/D', $v)</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with '&quot;1&quot; !== (string) $v'.">!preg_match('/^1$/D', $v)</warning>;
    return $r;
}

function reviewText(string $s)
{
    $r[] = <warning descr="Replace with 'false !== strpos($s, &quot;a-zA-Z&quot;)'.">preg_match('/a-zA-Z/', $s)</warning> ? 1 : 0;
    $r[] = <warning descr="Replace with 'str_replace(&quot;0-9a-zA-Z&quot;, '', $s)'.">preg_replace('/0-9a-zA-Z/', '', $s)</warning>;
    // trim() takes a list of bytes, not of characters
    $r[] = preg_replace(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/^é+/'</error>, '', $s);
    $r[] = preg_replace(<error descr="Non-ASCII characters in the pattern need the /u flag.">'/ +$/D'</error>, '', $s);
    $r[] = <warning descr="Replace with 'rtrim($s, ' ')'.">preg_replace('/ +$/D', '', $s)</warning>;
    return $r;
}

function reviewNamed(string $s, array $rest)
{
    $r[] = <warning descr="Replace with 'explode(&quot;,&quot;, $s, 2)'.">preg_split('/,/', $s, limit: 2)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;,&quot;, $s)'.">preg_split('/,/', subject: $s)</warning>;
    // flags is not the limit; a reordered or spread call keeps its shape
    $r[] = preg_split('/,/', $s, flags: 1);
    $r[] = preg_split(subject: $s, pattern: '/,/');
    $r[] = preg_split('/,/', ...$rest);
    $r[] = preg_match(subject: $s, pattern: <error descr="The /e flag was removed from PCRE; use a callback replacement.">'/tmp/e'</error>);
    $r[] = <warning descr="Pass the delimiter to preg_quote() so it is escaped as well.">preg_quote(str: '#a.b#')</warning>;
    return $r;
}

function reviewReports(string $s)
{
    preg_match(<error descr="'k' is not a valid PCRE modifier.">'/abc/i k'</error>, $s, $m);
    preg_match(<warning descr="Pattern has no valid delimiters.">"abc\x2f"</warning>, $s, $m);
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u).">'/\\\\[0-9]/'</weak_warning>, $s, $m); // an escaped backslash, then a class
    preg_match(<weak_warning descr="Write '[0-9]' as '\d' (same result without /u).">'/\[0-9][0-9]/'</weak_warning>, $s, $m);
    preg_match(<weak_warning descr="Collapse '[ab][ab]' into '[ab]' with a counted quantifier.">'/x\[ab][ab][ab]/'</weak_warning>, $s, $m);
}
