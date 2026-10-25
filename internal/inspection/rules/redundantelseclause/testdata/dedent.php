<?php
function profile(array $config) {
    if (!empty($config['profile'])) {
        return $config['profile'];
    } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> {
        // Fall back to settings.
        $p = getenv('PROFILE');
        if ($p === false) {
            $p = 'standard';
        }
// flush-left note
        return $p;
    }
}

function banner(bool $quiet) {
    if ($quiet) {
        return '';
    } <warning descr="Drop the 'else' and move its code after the 'if'.">else</warning> {
        $text = <<<TXT
        Welcome
        TXT;
        return $text;
    }
}
