<?php
// A class extending the receiver's, declared twice with different methods, may replace the callee.
class Reader { function take(string $s) {} }
if ($legacy) {
    class LegacyReader extends Reader {}
} else {
    class LegacyReader extends Reader { function take($s) {} }
}
function readInto($path, Reader $reader) { $reader->take(file_get_contents($path)); }
