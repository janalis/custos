<?php
function definite_replacement(string $path): void {
    $body = file_get_contents($path);
    if (true) { $body = 'ready'; }
    trim($body);
}
function skipped_replacement(string $path): void {
    $body = file_get_contents($path);
    if (false) { $body = 'ready'; }
    <warning descr="Handle file reading failure before using its result.">trim($body)</warning>;
}
function chosen_failure_arm(string $path): void {
    $body = true ? file_get_contents($path) : 'ready';
    <warning descr="Handle file reading failure before using its result.">trim($body)</warning>;
}
function skipped_failure_arm(string $path): void {
    $body = false ? file_get_contents($path) : 'ready';
    trim($body);
}
