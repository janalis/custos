<?php
namespace Store {
    use Countable as Sized;

    interface Readable {}
    interface Seekable extends Readable {}
    abstract class Stream implements Readable {}
    abstract class FileStream implements Seekable, \Countable {}
    abstract class Mid extends FileStream {}

    class Socket extends Stream
        implements
            <warning descr="'\Store\Readable' is already implemented by '\Store\Stream'; remove it here.">Readable</warning> {}

    class Disk extends FileStream implements <warning descr="'\Store\Readable' is already implemented by '\Store\FileStream'; remove it here.">Readable</warning>, <warning descr="'\Countable' is already implemented by '\Store\FileStream'; remove it here.">Sized</warning> {}

    class Tape extends FileStream implements \JsonSerializable, <warning descr="'\Store\Seekable' is already implemented by '\Store\FileStream'; remove it here.">Seekable</warning> {
        public function jsonSerialize(): mixed { return null; }
        public function count(): int { return 0; }
    }

    class Deep extends Mid implements <warning descr="'\Countable' is already implemented by '\Store\Mid'; remove it here.">\Countable</warning>, \Stringable {
        public function __toString(): string { return ''; }
    }

    $anon = new class extends Stream implements <warning descr="'\Store\Readable' is already implemented by '\Store\Stream'; remove it here.">Readable</warning> {};
}
