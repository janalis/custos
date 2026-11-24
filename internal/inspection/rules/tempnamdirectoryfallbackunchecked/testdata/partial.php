<?php
// The closure creates the temporary file on each call; no path is written yet.
function later($privateDir, $secret)
{
    $path = tempnam($privateDir, 'job', ...);
    file_put_contents($path, $secret);
}
