<?php
function temporary_options($path) {
    $stream = fopen($path, 'r');
    unknown_handler(['sink' => $stream]);
    <warning descr="Check that the stream opened successfully.">rewind($stream)</warning>;
}
function unknown_reference_parameter($path, $service) {
    $stream = fopen($path, 'r');
    $service->replace($stream);
    rewind($stream);
}
