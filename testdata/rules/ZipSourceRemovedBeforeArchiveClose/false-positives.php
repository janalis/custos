<?php
function guardReturn() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { if (!$z->close()) { return false; } unlink("source.txt"); }
    return true;
}
function guardBranch() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { if ($z->close()) { unlink("source.txt"); return true; } }
    return false;
}
function closeInTry() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { try { $z->close(); } finally { unlink("source.txt"); } }
}
function closeInFinally() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { try { audit(); } finally { $z->close(); } unlink("source.txt"); }
}
function deferred() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { $cleanup = function () { unlink("source.txt"); }; $z->close(); $cleanup(); }
}
function chained() {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { $z->close() && unlink("source.txt"); }
}
function elseifClose($skip) {
    $z = new ZipArchive(); $z->open("bundle.zip", ZipArchive::CREATE);
    if ($z->addFile("source.txt", "note.txt")) { if ($skip) { return; } elseif ($z->close()) { unlink("source.txt"); } }
}
final class ReviewClosedBundle {
    private ZipArchive $zip;
    public function finish() {
        if ($this->zip->addFile("source.txt", "note.txt")) { if (!$this->zip->close()) { return false; } unlink("source.txt"); }
        return true;
    }
}
