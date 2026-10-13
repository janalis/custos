<?php
$p=new XSLTProcessor();$r=$p->transformToXML($doc);if($r!==false){file_put_contents($out,$r);echo $r;}echo "x";function dead(){return;$p=new XSLTProcessor();echo $p->transformToXML($d);}

strlen("unrelated");

function truthyExit(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (!$out) { throw new RuntimeException('failed'); } $n = 1; echo $out; return $n; }
function truthyBranch(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); if ($out) { return file_put_contents($file, $out) !== false; } return false; }
function typedBranch(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); if (is_string($out)) { file_put_contents($file, $out); } }
function typedExit(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (!is_string(value: $out)) { return; } echo $out; }
function looseExit(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if ($out == false) { return; } echo $out; }
function looseFlipped(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (false != $out) { echo $out; } }
function emptyExit(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (empty($out)) { return; } echo $out; }
function elseBranch(XSLTProcessor $p, DOMDocument $d) { $out = $p->transformToXML($d); if (!$out) { echo 'failed'; } else { echo $out; } }
function chainElse(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x) { echo 'x'; } elseif (!$out) { echo 'failed'; } else { echo $out; } }
function chainBranch(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x) { echo 'x'; } elseif ($out) { echo $out; } }
function ternaryThen(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return $out ? file_put_contents($file, $out) : 0; }
function ternaryElse(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return !$out ? 0 : file_put_contents($file, $out); }
function shortAnd(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return $out && file_put_contents($file, $out); }
function shortOr(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); return !$out || file_put_contents($file, $out); }
function wordAnd(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); $out and file_put_contents($file, $out); }
function wordOr(XSLTProcessor $p, DOMDocument $d, $file) { $out = $p->transformToXML($d); empty($out) or file_put_contents($file, $out); }
function bothLeft(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($out && $x) { echo $out; } }
function bothRight(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x && ($out)) { echo $out; } }
function eitherLeft(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if (!$out || $x) { return; } echo $out; }
function eitherRight(XSLTProcessor $p, DOMDocument $d, $x) { $out = $p->transformToXML($d); if ($x || !$out) { return; } echo $out; }
function assigned(XSLTProcessor $p, DOMDocument $d) { if (!($out = $p->transformToXML($d))) { return; } echo $out; }

function retried_transformation(XSLTProcessor $processor, DOMDocument $document) {
    $attempts = 0;
    again:
    try {
        $markup = $processor->transformToXML($document);
        if (!is_string($markup)) {
            return 0;
        }
        echo $markup;
    } catch (RuntimeException $failure) {
        if (++$attempts < 3) {
            goto again;
        }
        throw $failure;
    }
    return 1;
}
