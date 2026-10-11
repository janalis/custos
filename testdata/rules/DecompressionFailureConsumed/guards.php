<?php
function silenced($b) { $s = @gzdecode($b); return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; }
function silencedDirect($b) { return <warning descr="Check decompression before consuming data.">strlen(@(gzdecode($b)))</warning>; }
function truthiness($b) { $s = gzdecode($b); return $s ? <warning descr="Check decompression before consuming data.">strlen($s)</warning> : 0; }
function wrongArm($b) { $s = gzdecode($b); return $s !== false ? 0 : <warning descr="Check decompression before consuming data.">strlen($s)</warning>; }
function wrongShortArm($b) { $s = gzdecode($b); return $s !== false ?: <warning descr="Check decompression before consuming data.">strlen($s)</warning>; }
function wrongOperator($b) { $s = gzdecode($b); return $s === false && <warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3; }
function leftOperand($b) { $s = gzdecode($b); return <warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3 && $s !== false; }
function looseTest($b) { $s = gzdecode($b); return $s != false && <warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3; }
function otherValue($b, $t) { $s = gzdecode($b); return $t !== false && $s !== null && <warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3; }
function reassigned($b) { $s = gzdecode($b); if ($s !== false) { $s = gzdecode($b); return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } return 0; }
function otherAssignment($b) { $s = gzdecode($b); if (($t = gzdecode($b)) !== false && ($s .= 'x') !== false) { return 1; } return 0; }
function staleAssignment($b) { if (($s = gzdecode($b)) !== false) { $s = gzdecode($b); return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } return 0; }
function mixedFailure($b, $x) { $s = gzdecode($b); if ($s === false && $x) { return 0; } return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; }
function mixedSuccess($b, $x) { $s = gzdecode($b); if ($s !== false || $x) { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } return 0; }
function negatedSuccess($b) { $s = gzdecode($b); if (!($s !== false)) { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } return 0; }
function bitwiseNot($b) { $s = gzdecode($b); if (~($s === false)) { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } return 0; }
function unrelatedLoops($b, $n) { $s = gzdecode($b); while ($n--) { for (;;) { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } } return 0; }
function inCondition($b) { $s = gzdecode($b); if (<warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3) { return 1; } return <warning descr="Check decompression before consuming data.">strlen($s)</warning> > 3 ? 1 : 0; }
function loggedOnly($b) { $s = gzdecode($b); if ($s === false) { error_log('bad'); } return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; }
function unrelatedArms($b, $x) { $s = gzdecode($b); if ($x) { return 0; } elseif ($b) { return 2; } else { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } }
function earlierArms($b, $x) { $s = gzdecode($b); if ($x) { return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; } elseif ($b) { return <warning descr="Check decompression before consuming data.">strtoupper($s)</warning>; } elseif ($s === false) { return 0; } return 1; }
function exitThenReassigned($b) { if (($s = gzdecode($b)) === false) { return 0; } $n = 1; $s = gzdecode($b); return <warning descr="Check decompression before consuming data.">strlen($s)</warning> + $n; }
function unbracedUse($b, $x) { $s = gzdecode($b); if ($x) return <warning descr="Check decompression before consuming data.">strlen($s)</warning>; return 0; }
