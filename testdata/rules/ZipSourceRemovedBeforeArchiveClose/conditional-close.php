<?php
function closeInOtherBranch($early) {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { if ($early) { $z->close(); } <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; $z->close(); }
}
function closeMayBeSkipped() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { try { $z->close(); } catch (Exception $e) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; } }
}
function otherArchive() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    $other = new ZipArchive(); $other->open("other.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { if ($other->close()) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; } $z->close(); }
}
function closeAfterProtectedCall() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { try { audit(); $z->close(); } finally { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; } }
}
function closeAfterSwallowedCall() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { try { audit(); $z->close(); } catch (Exception $e) { } <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; }
}
final class ReviewPendingBundle {
    private ZipArchive $zip;
    private ZipArchive $other;
    public function early() {
        if ($this->zip->addFile("source.txt", "note.txt")) { <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; $this->zip->close(); }
    }
    public function wrongArchive() {
        if ($this->zip->addFile("source.txt", "note.txt")) { $this->other->close(); <warning descr="Keep ZIP source files until the archive closes.">unlink("source.txt")</warning>; }
    }
}
