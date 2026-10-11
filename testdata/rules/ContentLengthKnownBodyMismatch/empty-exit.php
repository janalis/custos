<?php
function emptyExit() {
    <warning descr="Match Content-Length to the emitted bytes.">header("Content-Length: 2")</warning>;
    echo "hello";
    exit();
}
function emptyDie() {
    header("Content-Type: text/plain");
    <warning descr="Match Content-Length to the emitted bytes.">header("Content-Length: 2")</warning>;
    print "hel" . str_repeat("l", 1);
    echo "o", "!";
    die();
}
function matching() {
    header("Content-Length: 5");
    echo "hello";
    exit();
}
function unknownOutput($body) {
    header("Content-Length: 2");
    echo $body;
    die();
}
function noBody() {
    header("Content-Length: 5", true, 204);
    echo "hello!";
    exit();
}
