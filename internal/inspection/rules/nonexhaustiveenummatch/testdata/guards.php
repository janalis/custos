<?php
// Comparisons that do not exclude the missing case, or a subject that may change.
namespace GuardedSubject;

enum E { case A; case B; case C;
    const X = self::A;
    const Z = 3;
    function own(): int { if ($this === self::A) { return 0; } return <error descr="Handle every reachable enum case.">match($this) { self::B => 1 }</error>; }
}
/** @param-out E $r */
function byRef(E &$r): void { $r = E::C; }
function byVal(E $v, ...$rest): void {}
/** @param-out E $rest */
function refTail(E $v, &...$rest): void {}
function source(): E { return E::A; }
function bad01(E $e): int { if ($e === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B => 1 }</error>; }
function bad02(E $e): int { if ($e === E::A) { echo ''; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad03(E $e): int { if ($e === E::A) { return 0; } $e = E::A; return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad04(E $e, E $o): int { if ($o === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad05(E $e): int { if ($e === E::A || rand(0, 1) === 0) { return <error descr="Handle every reachable enum case.">match($e) { E::A => 1 }</error>; } return 0; }
function bad06(E $e): int { if ($e === E::A && rand(0, 1) === 0) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad07(E $e): int { if ($e === E::A) { return 0; } elseif (rand(0, 1) === 1) { return 1; } else { return <error descr="Handle every reachable enum case.">match($e) { E::C => 2 }</error>; } }
function bad08(E $e): int { return $e === E::B ? <error descr="Handle every reachable enum case.">match($e) { E::A => 1 }</error> : 0; }
function bad09(E $e): int { if ($e === E::A) { return 0; } elseif (rand(0, 1) === 1) { echo ''; } return <error descr="Handle every reachable enum case.">match($e) { E::A, E::C => 2 }</error>; }
function bad10(E $e): int { return <error descr="Handle every reachable enum case.">match($e) { E::X => 0, E::B => 1 }</error>; }
function bad11(E $e): int { if ($e === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match(source()) { E::B, E::C => 1 }</error>; }
function bad12(E $e): int { if (<error descr="Handle every reachable enum case.">match($e) { E::B, E::C => true }</error>) { return 1; } return 0; }
function bad13(E $e): int { return (<error descr="Handle every reachable enum case.">match($e) { E::B, E::C => true }</error>) ? 1 : 0; }
function bad14(E $e): int { return (int) (<error descr="Handle every reachable enum case.">match($e) { E::B, E::C => true }</error> || $e === E::A); }
function bad15(E $e): int { return (int) ($e === E::A xor <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => true }</error>); }
function bad16(E $e): int { if ($e === E::Z) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad17(E $e): int { if ($e == E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function bad18(E $e): int { if (-1 === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }

// Narrowing is withheld when the subject is not proven to keep its value.
function kept01(E $e): int { if ($e === E::A) { return 0; } (new \ArrayObject())->append($e); return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept02(E &$e): int { if ($e === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept03(E $e): int { if ($e === E::A) { return 0; } $r = &$e; return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept04(E $e): int { if ($e === E::A) { return 0; } $x = <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; $f = function () use (&$e) {}; return $x; }
function kept05(E $e): int { if ($e === E::A) { return 0; } $x = <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; $n = 'q'; $$n = 1; return $x; }
function kept06(E $e): int { if ($e === E::A) { return 0; } $x = <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; extract([]); return $x; }
function kept07(E $e): int { if ($e === E::A) { return 0; } $a = [&$e]; return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept08(E $e): int { if ($e === E::A) { return 0; } byRef($e); return <error descr="Handle every reachable enum case.">match($e) { E::A, E::B => 1 }</error>; }
function kept09(E $e): int { if ($e === E::A) { return 0; } refTail($e, $e, $e); return <error descr="Handle every reachable enum case.">match($e) { E::A, E::B => 1 }</error>; }
function kept10(E $e): int { if ($e === E::A) { goto done; } return 0; done: return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept11(E $e): int { if ($e === E::A) { return 0; } $x = <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; eval(''); return $x; }
function kept12(E $e): int { if ($e === E::A) { return 0; } $f = 'byVal'; $f($e); return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept13(E $e): int { if ($e === E::A) { return 0; } byVal(nope: $e); return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept14(E $e): int { if ($e === E::A) { return 0; } byVal(...$e); return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept15(E $e) { if ($e === E::A) { return 0; } yield $e; return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
function kept16(E $e): int { $f = function () use (&$e): int { /** @var E $e */ if ($e === E::A) { return 0; } return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }; return $f(); }
function kept17(E $e): int { if ($e === E::A) { return 0; } missing($e); return <error descr="Handle every reachable enum case.">match($e) { E::B, E::C => 1 }</error>; }
