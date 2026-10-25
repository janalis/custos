<?php
function profile(array $config) {
    if (!empty($config['profile'])) {
        return $config['profile'];
    }
    // Fall back to settings.
    $p = getenv('PROFILE');
    if ($p === false) {
        $p = 'standard';
    }
// flush-left note
    return $p;
}

function banner(bool $quiet) {
    if ($quiet) {
        return '';
    }
    $text = <<<TXT
        Welcome
        TXT;
        return $text;
}
