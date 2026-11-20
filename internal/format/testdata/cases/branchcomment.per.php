<?php
// presets: symfony per
if ($a) {
    foo();
    // describes the elseif
} elseif ($b) {
    bar();
    // ends the block
} else {
    baz();
}
try {
    foo();
    // describes the catch
} catch (E $e) {
    // Ignored.
} catch (F $e) {
}
