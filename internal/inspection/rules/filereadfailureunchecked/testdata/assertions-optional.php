<?php
function optional_failure_check(string $path): void {
    $contents = file_get_contents($path);
    assert($contents !== false);
    <warning descr="Handle file reading failure before using its result.">trim($contents)</warning>;
}
