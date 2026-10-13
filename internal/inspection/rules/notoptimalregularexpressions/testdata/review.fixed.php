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
        $r[] = false !== strpos($s, "tmp") ? 1 : 0;
        $r[] = false !== strpos($s, "tmp") ? 1 : 0;
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
    $old = '/tmp/e';
    $r[] = preg_match($old, $s) ? 1 : 0;
    return $r;
}

function reviewSubject(string $s, $v, array $q, bool $flag)
{
    $r[] = "a" === $s ? 1 : 0;
    $r[] = "a" === $s . 'a' ? 1 : 0;
    $r[] = "a" === ($flag ? 'zzz' : 'a') ? 1 : 0;
    $r[] = "a" === (string) ($q['k'] ?? 'a') ? 1 : 0;
    $r[] = "1" === (string) $v ? 1 : 0;
    $r[] = "1" !== (string) $v;
    return $r;
}

function reviewText(string $s)
{
    $r[] = false !== strpos($s, "a-zA-Z") ? 1 : 0;
    $r[] = str_replace("0-9a-zA-Z", '', $s);
    // trim() takes a list of bytes, not of characters
    $r[] = preg_replace('/^é+/', '', $s);
    $r[] = preg_replace('/ +$/D', '', $s);
    $r[] = rtrim($s, ' ');
    return $r;
}

function reviewNamed(string $s, array $rest)
{
    $r[] = explode(",", $s, 2);
    $r[] = explode(",", $s);
    // flags is not the limit; a reordered or spread call keeps its shape
    $r[] = preg_split('/,/', $s, flags: 1);
    $r[] = preg_split(subject: $s, pattern: '/,/');
    $r[] = preg_split('/,/', ...$rest);
    $r[] = preg_match(subject: $s, pattern: '/tmp/e');
    $r[] = preg_quote(str: '#a.b#');
    return $r;
}

function reviewReports(string $s)
{
    preg_match('/abc/i k', $s, $m);
    preg_match("abc\x2f", $s, $m);
    preg_match('/\\\\[0-9]/', $s, $m); // an escaped backslash, then a class
    preg_match('/\[0-9][0-9]/', $s, $m);
    preg_match('/x\[ab][ab][ab]/', $s, $m);
}
