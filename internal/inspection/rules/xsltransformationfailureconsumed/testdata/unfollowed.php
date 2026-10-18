<?php
// Code the scope evaluates leaves only the guards written around the use.
function thenGuard(XSLTProcessor $p, $doc, $out, $code) { $r = $p->transformToXML($doc); $r ? file_put_contents($out, $r) : 0; eval($code); }
function elseGuard(XSLTProcessor $p, $doc, $out, $code) { $r = $p->transformToXML($doc); !$r ? 0 : file_put_contents($out, $r); eval($code); }
function elseIfGuard(XSLTProcessor $p, $doc, $c, $code) { $r = $p->transformToXML($doc); if ($c) { echo 1; } elseif ($r) { echo $r; } eval($code); }
function andGuard(XSLTProcessor $p, $doc, $out, $code) { $r = $p->transformToXML($doc); $r && file_put_contents($out, $r); eval($code); }
