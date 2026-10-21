package infer_test

import "testing"

func TestOutParameters(t *testing.T) {
	check(t, `<?php
namespace App;
class Foo {}
/**
 * @param-out Foo $out
 * @param-out int $n
 */
function fill(&$out, &$n, &$plain, $byval) {}
class Svc {
    /** @phpstan-param-out list<string> $lines */
    public function read(?array &$lines): void {}
    /** @psalm-param-out Foo $x */
    public static function make(&$x): void {}
    public function run(string $s) {
        $m = null;
        preg_match('/a/', $s, $m);
        t('m', $m);
        preg_match('/a/', $s, $mf, PREG_OFFSET_CAPTURE); t('mf', $mf);
        preg_match_all('/a/', $s, $all); t('all', $all);
        if (preg_match('/(a)/', $s, $c)) { t('c', $c); }
        str_replace('a', 'b', $s, $count); t('count', $count);
        parse_str($s, $q); t('q', $q);
        exec('ls', $out, $code); t('out', $out); t('code', $code);
        $ints = [1]; exec('ls', $ints); t('ints', $ints);
        fill($o, $n, $p, $b); t('o', $o); t('n', $n);
        $p = 'x'; fill($o2, $n2, $p, $b); t('p', $p);
        $this->read($lines); t('lines', $lines);
        self::make($x); t('x', $x);
        $y = 1; $s !== '' && preg_match('/a/', $s, $y); t('y', $y);
    }
}
`, map[string]string{
		"m": "string[]", "mf": "array[]", "all": "string[][]", "c": "string[]", "count": "int",
		"q": "array", "out": "string[]", "code": "int", "ints": "int[]|string[]",
		"o": `\App\Foo`, "n": "int", "p": "?unknown", "lines": "string[]", "x": `\App\Foo`,
		"y": "int|string[]",
	})
}

func TestPregMatchOutputTypes(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $s, $flags) {
    preg_match('/(a)(b)?/', $s, $m);
    t('m', $m);
    preg_match('/(a)(b)?/', $s, $m2, PREG_UNMATCHED_AS_NULL);
    t('m2', $m2);
    preg_match_all('/(a)(b)?/', $s, $m3, PREG_UNMATCHED_AS_NULL);
    t('m3', $m3);
    preg_match('/(a)/', $s, $m4, 0);
    t('m4', $m4);
    preg_match('/(a)/', $s, $m5, PREG_OFFSET_CAPTURE);
    t('m5', $m5);
    preg_match('/(a)/', $s, $m6, $flags);
    t('m6', $m6);
    preg_match('/(a)/', $s, $m7, ...[0]);
    t('m7', $m7);
}
`, map[string]string{
		"m":  "string[]",
		"m2": "null[]|string[]",
		"m3": "null[][]|string[][]",
		"m4": "string[]",
		"m5": "array[]<array{0: string, 1: int}>",
		"m6": "array",
		"m7": "array",
	})
}

// A PHP function without a listed output contract leaves in the variable
// what its parameter is declared to take; moving the pointer of an array or
// removing one of its ends keeps the array's type.
func TestUnlistedBuiltinOutParameters(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $host, $pid, array $items) {
    $hosts = 'x'; getmxrr($host, $hosts); t('hosts', $hosts);
    $status = 'x'; pcntl_waitpid($pid, $status); t('status', $status);
    $list = [1, 2]; end($list); t('list', $list);
    $popped = [1, 2]; array_pop($popped); t('popped', $popped);
    $walked = [1, 2]; array_walk($walked, 'f'); t('walked', $walked);
    $object = new ArrayObject(); array_walk($object, 'f'); t('object', $object);
    $getmxrrHosts = 1; getmxrr($host, $getmxrrHosts, $weights); t('weights', $weights);
    $read = 'x'; stream_select($read, $write, $except, 0); t('read', $read);
    $code = 'x'; msg_receive($queue, 0, $type, 1, $message, true, 0, $code); t('code', $code);
}
`, map[string]string{"hosts": "?unknown", "status": "?unknown", "list": "int[]{0: int, 1: int}", "popped": "int[]", "walked": "array", "object": "?unknown", "weights": "?unknown", "read": "array|null", "code": "int"})
}

// preg_match_all() lists strings per group or, with PREG_SET_ORDER, per
// match; PREG_OFFSET_CAPTURE pairs each match with its byte offset.
func TestPregMatchFlagShapes(t *testing.T) {
	checkAnywhere(t, `<?php
function f(string $s) {
    preg_match_all('/(a)(b)/', $s, $set, PREG_SET_ORDER);
    foreach ($set as $m) { t('setCell', $m[2]); }
    preg_match_all('/(a)(b)/', $s, $groups, PREG_PATTERN_ORDER);
    t('groups', $groups);
    preg_match_all('/(a)/', $s, $all, PREG_OFFSET_CAPTURE);
    foreach ($all[1] as $pair) { t('text', $pair[0]); t('offset', $pair[1]); }
    preg_match('/(a)/', $s, $one, PREG_OFFSET_CAPTURE);
    t('one', $one[1][1]);
    preg_match('/(a)/', $s, $invalid, PREG_SET_ORDER);
    t('invalid', $invalid);
}
`, map[string]string{"setCell": "string", "groups": "string[][]", "text": "string", "offset": "int", "one": "int", "invalid": "array"})
}
