<?php
// Prints PHP's own token stream for each file given on stdin (one path per
// line) as "NAME START END" lines, files separated by "== path".
while (($path = fgets(STDIN)) !== false) {
    $path = rtrim($path, "\n");
    echo "== $path\n";
    $pos = 0;
    foreach (token_get_all(file_get_contents($path)) as $t) {
        if (is_array($t)) {
            $len = strlen($t[1]);
            echo token_name($t[0]), ' ', $pos, ' ', $pos + $len, "\n";
        } else {
            $len = strlen($t);
            echo 'CHAR:', $t, ' ', $pos, ' ', $pos + $len, "\n";
        }
        $pos += $len;
    }
}
