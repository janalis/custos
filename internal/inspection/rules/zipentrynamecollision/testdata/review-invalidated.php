<?php
function renamed(string $dir) {
    $z = new ZipArchive(); $z->open("$dir/a.zip", ZipArchive::CREATE);
    if ($z->addFromString("n.txt", "one")) { $z->renameName("n.txt", "old.txt"); $z->addFromString("n.txt", "two"); }
}
function rebound(string $dir) {
    $z = new ZipArchive(); $z->open("$dir/b.zip", ZipArchive::CREATE);
    if ($z->addFromString("n.txt", "one")) { $first = $z; $z = new ZipArchive(); $z->open("$dir/c.zip", ZipArchive::CREATE); $z->addFromString("n.txt", "two"); }
}
function deleted(string $dir) {
    $z = new ZipArchive(); $z->open("$dir/d.zip", ZipArchive::CREATE);
    if ($z->addFromString("n.txt", "one")) { $z->deleteIndex(0); $z->addFromString("n.txt", "two"); }
}
function sometimesDeleted(string $dir, bool $drop) {
    $z = new ZipArchive(); $z->open("$dir/e.zip", ZipArchive::CREATE);
    if ($z->addFromString("n.txt", "one")) { if ($drop) { $z->deleteIndex(0); } <warning descr="Make ZIP entry replacement explicit.">$z->addFromString("n.txt", "two")</warning>; }
}
function reopenedFirst(string $dir) {
    $z = new ZipArchive(); $z->open("$dir/f.zip", ZipArchive::CREATE); $z->close(); $z->open("$dir/f.zip");
    if ($z->addFromString("n.txt", "one")) { <warning descr="Make ZIP entry replacement explicit.">$z->addFromString("n.txt", "two")</warning>; }
}
final class ReviewBundle {
    private ZipArchive $zip;
    private int $count = 0;
    public function collide() {
        if ($this->zip->addFromString("n.txt", "one")) { $this->count = 1; <warning descr="Make ZIP entry replacement explicit.">$this->zip->addFromString("n.txt", "two")</warning>; }
    }
    public function dropped() {
        if ($this->zip->addFromString("n.txt", "one")) { $this->zip->deleteName("n.txt"); $this->zip->addFromString("n.txt", "two"); }
    }
    public function replaced(ZipArchive $next) {
        if ($this->zip->addFromString("n.txt", "one")) { $this->zip = $next; $this->zip->addFromString("n.txt", "two"); }
    }
}
