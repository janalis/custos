<?php
function released_binding($items, $path) {
    foreach ($items as &$value) {}
    unset($value);
    $value = file_get_contents($path);
    unrelated_operation();
    <warning descr="Handle file reading failure before using its result.">trim($value)</warning>;
}
function existing_binding($items, $path) {
    foreach ($items as &$value) {}
    $value = file_get_contents($path);
    unrelated_operation();
    trim($value);
}
function unknown_method_parameter($service, $path) {
    $value = file_get_contents($path);
    $service->replace($value);
    trim($value);
}
function reference_only_on_exiting_path($path, $bad) {
    $value = file_get_contents($path);
    if ($bad) { throw new UnknownFailureContext($value); }
    <warning descr="Handle file reading failure before using its result.">trim($value)</warning>;
}
