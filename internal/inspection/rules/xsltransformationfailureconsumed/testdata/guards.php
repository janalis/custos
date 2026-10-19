<?php
function otherValue(XSLTProcessor $p, DOMDocument $d, $other) { $out = $p->transformToXML($d); if ($other) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function reassigned(XSLTProcessor $p, DOMDocument $d, DOMDocument $e) { $out = $p->transformToXML($d); if (!$out) { return; } $out = $p->transformToXML($e); <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> }
function failingBranch(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (empty($out)) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function notTerminating(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (!$out) { echo 'failed'; } <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> }
function guardWithElse(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (!$out) { return; } else { $n = 1; } echo $out; }
function unrelatedChain(XSLTProcessor $p, DOMDocument $d, $x, $y) { $out = $p->transformToXML($d); if ($x) { echo 'x'; } elseif ($y) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } else { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function shortTernary(XSLTProcessor $p, DOMDocument $d, $file, $x) { $out = $p->transformToXML($d); return $x ?: <warning descr="Reject transformation failure before using the XML.">file_put_contents($file, $out)</warning>; }
function leftOperand(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return <warning descr="Reject transformation failure before using the XML.">file_put_contents($file, $out)</warning> && $out; }
function arithmetic(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return 1 + <warning descr="Reject transformation failure before using the XML.">file_put_contents($file, $out)</warning>; }
function silenced(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (@$out) { echo $out; } }
function otherTest(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (is_scalar($out)) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function eitherPasses(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($out || $x) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function bothFail(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if (!$out && $x) { return; } <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> }
function comparesTrue(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out == true) { return; } <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> }
function notNull(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out !== null) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function ordered(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out < 'a') { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function otherAssignment(XSLTProcessor $p, DOMDocument $d) { if ($first = $p->transformToXML($d)) { <warning descr="Reject transformation failure before using the XML.">echo $p->transformToXML($d);</warning> } }
function compound(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x .= 'a') { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function bothOperandsReject(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out || is_string($out)) { echo $out; } }
function bothOperandsFail(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out === false && !is_string($out)) { return; } echo $out; }
function oneOperandRejects(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x || is_string($out)) { <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> } }
function oneOperandFails(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($out === false && $x) { return; } <warning descr="Reject transformation failure before using the XML.">echo $out;</warning> }
