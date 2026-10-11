<?php
function reviewPreIncrement(ZipArchive $z) { $i = -1; while (++$i < $z->numFiles) { echo $z->getNameIndex($i); } }
function reviewPostIncrement(ZipArchive $z) { $i = -1; do { $i++; $n = $z->getNameIndex($i); } while ($n !== false); }
function reviewStatement(ZipArchive $z) { $i = -1; $i++; echo $z->statIndex($i); }
function reviewDecrement(ZipArchive $z) { $i = -1; $i--; $i += 3; echo $z->getFromIndex($i); }
function reviewCopy(ZipArchive $z) { $i = -1; ++$i; $j = $i; echo $z->getNameIndex(($j)); }
function reviewCopyCycle(ZipArchive $z) { $i = -1; $j = $i; $i = $j; $j++; echo $z->getNameIndex($i); }
function reviewReference(ZipArchive $z) { $i = -1; $bump = function () use (&$i) { $i = 0; }; $bump(); echo $z->getNameIndex($i); }
function reviewArrow(ZipArchive $z) { return fn() => [$i = -1, $i++, $z->getNameIndex($i)]; }
function reviewArrowOuter(ZipArchive $z) { return fn() => [$i = -1, fn() => [$i++, $z->getNameIndex($i)]]; }
final class ReviewCursor {
    private int $position = -1;
    private static int $shared = -1;
    public function __construct(private ZipArchive $zip) {}
    public function next() { $this->position++; return $this->zip->getNameIndex($this->position); }
    public function shared() { self::$shared++; return $this->zip->getNameIndex(self::$shared); }
}
