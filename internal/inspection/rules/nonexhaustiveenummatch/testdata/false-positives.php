<?php
// Strict comparisons against cases narrow a subject that keeps its value.
namespace NarrowedSubject;

enum E { case A; case B; case C;
    const X = self::A;
    const Y = E::X;
    const Q = \NarrowedSubject\E::B;
    const Z = 3;
    const P = self::R;
    const R = self::P;
    const W = Other::A;
    function own(): int { if ($this === self::A) { return 0; } return match($this) { self::B, self::C => 1 }; }
}
enum Other { case A; }
function byVal(E $v, ...$rest): void {}
function none(): void {}
function ok01(E $e): int { if ($e === E::A) { return 0; } return match($e) { E::B, E::C => 1 }; }
function ok02(E $e): int { if ($e !== E::A) { return match($e) { E::B, E::C => 1 }; } return 0; }
function ok03(E $e): int { return $e === E::B ? 0 : match($e) { E::A, E::C => 1 }; }
function ok04(E $e): int { return $e === E::B ? match($e) { E::B => 1 } : 0; }
function ok05(E $e): int { if ($e === E::A) { return 0; } elseif ($e === E::B) { return 1; } else { return match($e) { E::C => 2 }; } }
function ok06(E $e): int { if ($e === E::A) { return 0; } elseif ($e !== E::B) { return match($e) { E::C => 2 }; } return 3; }
function ok07(E $e): int { if (E::A === $e || $e === E::B) { return 0; } return match($e) { E::C => 2 }; }
function ok08(E $e): int { if (!($e === E::A) && $e !== E::B) { return match($e) { E::C => 2 }; } return 0; }
function ok09(E $e): int { return (int) ($e === E::A || match($e) { E::B, E::C => true }); }
function ok10(E $e): int { return (int) ($e !== E::A and match($e) { E::B, E::C => true }); }
function ok11(E $e): int { if ($e === E::X) { return 0; } return match($e) { E::B, E::C => 1 }; }
function ok12(E $e): int { return match($e) { E::Y => 0, E::Q, E::C => 1 }; }
function ok13(E $e): int { foreach ([1, 2] as $i) { if ($e === E::A) { continue; } echo match($e) { E::B, E::C => '' }; } return 0; }
function ok14(E $e): int { if ($e === E::A) { throw new \LogicException(); } byVal($e, 1, $e); return match($e) { E::B, E::C => 1 }; }
function ok15(E $e): int { if ($e === E::A) { return 0; } elseif ($e === E::B) { return 1; } return match($e) { E::C => 2 }; }
function ok16(E $e): int { $f = fn (): int => $e === E::A ? 0 : match($e) { E::B, E::C => 1 }; return $f(); }
function ok17(E $e): int { $f = function () use ($e): int { if ($e === E::A) { return 0; } return match($e) { E::B, E::C => 1 }; }; return $f(); }
function ok18(E $e): int { if ($e === E::A) { return 0; } if ($e === E::B) { return match($e) { E::B => 1 }; } return 2; }
function ok19(E $e): int { return $e === E::A ?: match($e) { E::B, E::C => 1 }; }
function ok20(E $e): int { if ($e === E::A) { return 0; } byVal(v: $e); none($e); return match($e) { E::B, E::C => 1 }; }
function ok21(E $e): int { if ($e === E::A) { return 0; } elseif (match($e) { E::B, E::C => true }) { return 1; } return 2; }
function ok22(E $e): int { if ($e === E::B or $e === E::C) { return 0; } return match($e) { E::X => 1 }; }
function ok23(E $e): int { $e = E::X; return match($e) { E::A => 1 }; }

// An arm whose constant cannot be traced to a case leaves the arm set unknown.
function unknown01(E $e): int { return match($e) { E::Z => 0, E::B => 1 }; }
function unknown02(E $e): int { return match($e) { E::P => 0, E::B => 1 }; }
function unknown03(E $e): int { return match($e) { E::W => 0, E::B => 1 }; }
function unknown04(E $e): int { return match($e) { E::Missing => 0, E::B => 1 }; }
function unknown05(E $e, string $n): int { return match($e) { E::{$n} => 0, E::B => 1 }; }
