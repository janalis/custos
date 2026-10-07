<?php
namespace Store {
    interface Readable {}
    interface Seekable extends Readable {}
    abstract class Stream implements Readable {}

    class Pipe extends Stream implements Seekable {}
    class Loose implements Readable, Seekable {}
    class Ghost extends Stream implements Unknown {}
    class Orphan extends MissingParent implements Readable {}
}
