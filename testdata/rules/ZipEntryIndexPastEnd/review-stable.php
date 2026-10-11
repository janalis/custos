<?php
const REVIEW_NEGATIVE = -1;
function reviewConstant(ZipArchive $z) { echo <warning descr="Use an existing ZIP entry index.">$z->getNameIndex(REVIEW_NEGATIVE)</warning>; }
function reviewAlias(ZipArchive $z) { $i = -1; $j = $i; $i = $j; echo <warning descr="Use an existing ZIP entry index.">$z->deleteIndex($i)</warning>; }
function reviewOther(ZipArchive $z) { $i = -1; $k = 0; $k++; echo <warning descr="Use an existing ZIP entry index.">$z->getStreamIndex($i)</warning>; }
function reviewValueCapture(ZipArchive $z) { $i = -1; $f = function () use ($i) { return $i + 1; }; echo <warning descr="Use an existing ZIP entry index.">$z->getNameIndex($i)</warning>; }
function reviewArrowLocal(ZipArchive $z) { return fn() => [$i = -1, <warning descr="Use an existing ZIP entry index.">$z->getNameIndex($i)</warning>]; }
