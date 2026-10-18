<?php
// An unpacked list may name the stream: whether it was closed is unknown.
function closedBySpread(array $handles) { $h = fopen("f", "r"); if ($h === false) { return; } fclose(...$handles); fread($h, 1); }
