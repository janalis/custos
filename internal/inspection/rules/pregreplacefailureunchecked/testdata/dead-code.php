<?php
function disabled(string $code): string
{
    $code = preg_replace('/a/', 'b', $code);
    if (0 && strpos($code, '{') === 0) {
        $code = "{\n" . substr($code, 1);
    }
    if (false) {
        echo strlen($code);
    }
    echo 1 || trim($code);
    return <warning descr="Handle regex replacement failure before using its result.">trim($code)</warning>;
}
