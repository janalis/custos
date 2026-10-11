<?php
function fillSpec(&$spec) { $spec = 'P1D'; }
function fillLast($first, &...$rest) { $rest[0] = 'P1D'; }
function show($spec, $extra = null) { echo $spec; }

class SpecSource
{
    public function load(&$into) { $into = 'P2D'; }
    public static function read(string $label, string &$into): void { $into = 'P2D'; }
}

// A parameter default is only one of the values callers pass.
function ttl(string $spec = '') {
    if ($spec === '') {
        return null;
    }
    return new DateInterval($spec);
}
function ttlMethodLike(string $spec = 'none', $other = 'P1D') {
    return $spec === 'none' ? null : new DateInterval($spec);
}
function fromParamChain($given) {
    $spec = $given ?: 'never';
    return new DateInterval($spec);
}
function fromParamCopy($given = 'bogus') {
    $spec = $given;
    return new DateInterval($spec);
}
function fromCoalesce(?string $given) {
    return new DateInterval($given ?? 'none');
}
function fromTernary($flag, $given) {
    return new DateInterval($flag ? 'none' : $given);
}
function captured($spec) {
    return function () use ($spec, $unused) {
        if ($spec === null) {
            $spec = 'default';
        }
        return new DateInterval($spec);
    };
}
function arrow() {
    $spec = 'x';
    return fn () => new DateInterval($spec);
}
function neverAssigned() {
    $copy = $missing ?: 'x';
    return new DateInterval($copy);
}

// Writes value discovery does not follow.
function viaRef() { $s = 'x'; fillSpec($s); return new DateInterval($s); }
function viaNamedRef() { $s = 'x'; SpecSource::read(label: 'a', into: $s); return new DateInterval($s); }
function viaMethodRef(SpecSource $src) { $s = 'x'; $src->load($s); return new DateInterval($s); }
function viaVariadicRef() { $s = 'x'; fillLast(1, 2, $s); return new DateInterval($s); }
function viaBuiltinRef() { $s = 'x'; preg_match('/P\d+D/', 'P1D', $s); return new DateInterval($s); }
function viaCallTimeRef() { $s = 'x'; unknownFiller(&$s); return new DateInterval($s); }
function viaForeach() { $s = 'x'; foreach (['P1D', 'P2D'] as $s) { } return new DateInterval($s); }
function viaForeachKey() { $s = 'x'; foreach (['P1D' => 1] as $s => $n) { } return new DateInterval($s); }
function viaAlias() { $s = 'x'; $r = &$s; $r = 'P3D'; return new DateInterval($s); }
function viaAliasTarget(array $all) { $s = 'x'; $s = &$all[0]; return new DateInterval($s); }
function viaArrayRef() { $s = 'x'; $slots = [&$s]; $slots[0] = 'P1D'; return new DateInterval($s); }
function viaClosureRef() { $s = 'x'; $set = function () use (&$s) { $s = 'P1D'; }; $set(); return new DateInterval($s); }
function viaDestructuring(array $row) { $s = 'x'; [$s, $n] = $row; return new DateInterval($s); }
function viaList(array $row) { $s = 'x'; list('spec' => $s) = $row; return new DateInterval($s); }
function viaGlobal() { $s = 'x'; global $s; return new DateInterval($s); }
function viaStatic() { static $s = 'x'; $s = 'y'; return new DateInterval($s); }
function viaExtract(array $o) { $s = 'x'; extract($o); return new DateInterval($s); }
function viaOffset() { $s = 'x1D'; $s[0] = 'P'; return new DateInterval($s); }
function viaCatch() { $s = 'x'; try { run(); } catch (Exception $s) { } return new DateInterval($s); }
function viaChain($given) { $base = 'x'; fillSpec($base); $s = $base; return new DateInterval($s); }
