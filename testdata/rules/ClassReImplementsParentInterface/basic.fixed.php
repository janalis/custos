<?php
namespace Store {
    use Countable as Sized;

    interface Readable {}
    interface Seekable extends Readable {}
    abstract class Stream implements Readable {}
    abstract class FileStream implements Seekable, \Countable {}
    abstract class Mid extends FileStream {}

    class Socket extends Stream {}

    class Disk extends FileStream {}

    class Tape extends FileStream implements \JsonSerializable {
        public function jsonSerialize(): mixed { return null; }
        public function count(): int { return 0; }
    }

    class Deep extends Mid implements \Stringable {
        public function __toString(): string { return ''; }
    }

    $anon = new class extends Stream {};
}
